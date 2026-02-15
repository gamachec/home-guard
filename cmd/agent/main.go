package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"home-guard/internal/config"
	"home-guard/internal/notify"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			installService()
			return
		case "uninstall":
			uninstallService()
			return
		case "notify":
			runNotify()
			return
		}
	}

	runAgent()
}

func runNotify() {
	if len(os.Args) < 3 {
		return
	}
	n, err := notify.DecodeNotification(os.Args[2])
	if err != nil {
		log.Fatalf("notify: invalid payload: %v", err)
	}
	if err := notify.NewWindowsNotifier().Send(n); err != nil {
		log.Fatalf("notify: send failed: %v", err)
	}
}

func writeVersionFile(dir, v string) {
	if err := os.WriteFile(filepath.Join(dir, "version.txt"), []byte(v), 0644); err != nil {
		log.Printf("failed to write version.txt: %v", err)
	}
}

func runAgent() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	execDir := filepath.Dir(execPath)
	configPath := filepath.Join(execDir, "config.json")

	writeVersionFile(execDir, version)

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := NewApp(cfg, configPath, notify.NewWindowsNotifier(), version)
	if err := app.Start(ctx); err != nil {
		log.Fatalf("failed to start: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	cancel()
	app.Stop()
}
