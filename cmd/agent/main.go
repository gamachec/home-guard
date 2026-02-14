package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"home-guard/internal/config"
	"home-guard/internal/mqtt"
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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	_ = client.PublishStatus("offline")
	client.Disconnect()
}
