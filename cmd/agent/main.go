package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"home-guard/internal/agent"
	"home-guard/internal/config"
	"home-guard/internal/mqtt"
	"home-guard/internal/notify"
	"home-guard/internal/process"
)

func main() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	configPath := filepath.Join(filepath.Dir(execPath), "config.json")

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	mqttClient := mqtt.NewClient(cfg)
	if err := mqttClient.Connect(); err != nil {
		log.Fatalf("failed to connect to MQTT broker: %v", err)
	}

	if err := mqttClient.PublishStatus("online"); err != nil {
		log.Printf("failed to publish status: %v", err)
	}

	manager := process.NewManager(process.NewWindowsAdapter())
	notifier := notify.NewWindowsNotifier()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	onPublish := func(mode agent.Mode) {
		statTopic := fmt.Sprintf("stat/%s/current_mode", cfg.ClientID)
		if err := mqttClient.Publish(statTopic, string(mode)); err != nil {
			log.Printf("failed to publish mode: %v", err)
		}
	}

	a := agent.New(manager, cfg, configPath, onPublish)

	recoverCh := make(chan agent.Mode, 1)
	var recoverOnce sync.Once

	statModeTopic := fmt.Sprintf("stat/%s/current_mode", cfg.ClientID)
	if err := mqttClient.Subscribe(statModeTopic, func(payload []byte) {
		recoverOnce.Do(func() {
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
		a.SetMode(ctx, mode)
	case <-time.After(2 * time.Second):
	}

	notifyTopic := fmt.Sprintf("cmnd/%s/notify", cfg.ClientID)
	if err := mqttClient.Subscribe(notifyTopic, func(payload []byte) {
		go func() {
			var n notify.Notification
			if err := json.Unmarshal(payload, &n); err != nil {
				n = notify.Notification{Title: "Home Guard", Message: strings.TrimSpace(string(payload))}
			}
			if err := notifier.Send(n); err != nil {
				log.Printf("failed to send notification: %v", err)
			}
		}()
	}); err != nil {
		log.Printf("failed to subscribe to %s: %v", notifyTopic, err)
	}

	modeTopic := fmt.Sprintf("cmnd/%s/mode", cfg.ClientID)
	if err := mqttClient.Subscribe(modeTopic, func(payload []byte) {
		mode := agent.Mode(strings.TrimSpace(string(payload)))
		a.SetMode(ctx, mode)
	}); err != nil {
		log.Printf("failed to subscribe to %s: %v", modeTopic, err)
	}

	blacklistTopic := fmt.Sprintf("cmnd/%s/blacklist/set", cfg.ClientID)
	if err := mqttClient.Subscribe(blacklistTopic, func(payload []byte) {
		var apps []string
		if err := json.Unmarshal(payload, &apps); err != nil {
			log.Printf("invalid blacklist payload: %v", err)
			return
		}
		if err := a.SetBlacklist(apps); err != nil {
			log.Printf("failed to save blacklist: %v", err)
		}
	}); err != nil {
		log.Printf("failed to subscribe to %s: %v", blacklistTopic, err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	cancel()
	_ = mqttClient.PublishStatus("offline")
	mqttClient.Disconnect()
}
