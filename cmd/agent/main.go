package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

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

	cfg, err := config.Load(filepath.Join(filepath.Dir(execPath), "config.json"))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client := mqtt.NewClient(cfg)
	if err := client.Connect(); err != nil {
		log.Fatalf("failed to connect to MQTT broker: %v", err)
	}

	if err := client.PublishStatus("online"); err != nil {
		log.Printf("failed to publish status: %v", err)
	}

	manager := process.NewManager(process.NewWindowsAdapter())
	notifier := notify.NewWindowsNotifier()

	notifyTopic := fmt.Sprintf("cmnd/%s/notify", cfg.ClientID)
	if err := client.Subscribe(notifyTopic, func(payload []byte) {
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

	killTopic := fmt.Sprintf("cmnd/%s/kill_test", cfg.ClientID)
	if err := client.Subscribe(killTopic, func(payload []byte) {
		name := strings.TrimSpace(string(payload))
		go func() {
			log.Printf("killing process: %s", name)
			if err := manager.KillByName(name); err != nil {
				log.Printf("failed to kill %s: %v", name, err)
			}
		}()
	}); err != nil {
		log.Printf("failed to subscribe to %s: %v", killTopic, err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	_ = client.PublishStatus("offline")
	client.Disconnect()
}
