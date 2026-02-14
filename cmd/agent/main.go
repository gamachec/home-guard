package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"home-guard/internal/config"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			installService()
			return
		case "uninstall":
			uninstallService()
			return
		}
	}

	if isWindowsService() {
		runAsService()
		return
	}

	runAgent()
}

func runAgent() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	configPath := filepath.Join(filepath.Dir(execPath), "config.json")

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := NewApp(cfg, configPath)
	if err := app.Start(ctx); err != nil {
		log.Fatalf("failed to start: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	cancel()
	app.Stop()
}
