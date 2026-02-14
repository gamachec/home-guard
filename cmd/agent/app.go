package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"home-guard/internal/agent"
	"home-guard/internal/config"
	"home-guard/internal/mqtt"
	"home-guard/internal/notify"
	"home-guard/internal/process"
)

type App struct {
	cfg        *config.Config
	configPath string
	mqtt       *mqtt.Client
	agent      *agent.Agent
	notifier   notify.Notifier
}

func NewApp(cfg *config.Config, configPath string) *App {
	manager := process.NewManager(process.NewWindowsAdapter())
	notifier := notify.NewWindowsNotifier()
	mqttClient := mqtt.NewClient(cfg)

	a := &App{
		cfg:        cfg,
		configPath: configPath,
		mqtt:       mqttClient,
		notifier:   notifier,
	}

	onPublish := func(mode agent.Mode) {
		statTopic := fmt.Sprintf("stat/%s/current_mode", cfg.ClientID)
		if err := mqttClient.Publish(statTopic, string(mode)); err != nil {
			log.Printf("failed to publish mode: %v", err)
		}
	}

	a.agent = agent.New(manager, cfg, configPath, onPublish)
	return a
}

func (a *App) Start(ctx context.Context) error {
	if err := a.mqtt.Connect(); err != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", err)
	}

	if err := a.mqtt.PublishStatus("online"); err != nil {
		log.Printf("failed to publish status: %v", err)
	}

	a.recoverMode(ctx)
	a.subscribeTopics(ctx)
	return nil
}

func (a *App) Stop() {
	_ = a.mqtt.PublishStatus("offline")
	a.mqtt.Disconnect()
}

func (a *App) recoverMode(ctx context.Context) {
	recoverCh := make(chan agent.Mode, 1)
	var once sync.Once

	statModeTopic := fmt.Sprintf("stat/%s/current_mode", a.cfg.ClientID)
	if err := a.mqtt.Subscribe(statModeTopic, func(payload []byte) {
		once.Do(func() {
			if mode := agent.Mode(strings.TrimSpace(string(payload))); mode != "" {
				recoverCh <- mode
			}
		})
	}); err != nil {
		log.Printf("failed to subscribe to %s: %v", statModeTopic, err)
	}

	select {
	case mode := <-recoverCh:
		log.Printf("recovered mode: %s", mode)
		a.agent.SetMode(ctx, mode)
	case <-time.After(2 * time.Second):
	}
}

func (a *App) subscribeTopics(ctx context.Context) {
	notifyTopic := fmt.Sprintf("cmnd/%s/notify", a.cfg.ClientID)
	if err := a.mqtt.Subscribe(notifyTopic, func(payload []byte) {
		go a.handleNotify(payload)
	}); err != nil {
		log.Printf("failed to subscribe to %s: %v", notifyTopic, err)
	}

	modeTopic := fmt.Sprintf("cmnd/%s/mode", a.cfg.ClientID)
	if err := a.mqtt.Subscribe(modeTopic, func(payload []byte) {
		a.agent.SetMode(ctx, agent.Mode(strings.TrimSpace(string(payload))))
	}); err != nil {
		log.Printf("failed to subscribe to %s: %v", modeTopic, err)
	}

	blacklistTopic := fmt.Sprintf("cmnd/%s/blacklist/set", a.cfg.ClientID)
	if err := a.mqtt.Subscribe(blacklistTopic, func(payload []byte) {
		a.handleBlacklist(payload)
	}); err != nil {
		log.Printf("failed to subscribe to %s: %v", blacklistTopic, err)
	}
}

func (a *App) handleNotify(payload []byte) {
	var n notify.Notification
	if err := json.Unmarshal(payload, &n); err != nil {
		n = notify.Notification{Title: "Home Guard", Message: strings.TrimSpace(string(payload))}
	}
	if err := a.notifier.Send(n); err != nil {
		log.Printf("failed to send notification: %v", err)
	}
}

func (a *App) handleBlacklist(payload []byte) {
	var apps []string
	if err := json.Unmarshal(payload, &apps); err != nil {
		log.Printf("invalid blacklist payload: %v", err)
		return
	}
	if err := a.agent.SetBlacklist(apps); err != nil {
		log.Printf("failed to save blacklist: %v", err)
	}
}
