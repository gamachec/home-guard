//go:build windows

package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"home-guard/internal/config"
	"home-guard/internal/notify"
)

const serviceName = "HomeGuard"

type windowsService struct{}

func (ws *windowsService) Execute(_ []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("service: panic: %v", p)
		}
	}()

	log.Printf("service: starting")
	s <- svc.Status{State: svc.StartPending}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	execPath, err := os.Executable()
	if err != nil {
		log.Printf("service: failed to get executable path: %v", err)
		return false, 1
	}

	configPath := filepath.Join(filepath.Dir(execPath), "config.json")
	log.Printf("service: loading config from %s", configPath)

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Printf("service: failed to load config: %v", err)
		return false, 1
	}

	app := NewApp(cfg, configPath, notify.NewSessionNotifier())
	if err := app.Start(ctx); err != nil {
		log.Printf("service: failed to start app: %v", err)
		return false, 1
	}

	log.Printf("service: running")
	s <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for req := range r {
		switch req.Cmd {
		case svc.Stop, svc.Shutdown:
			log.Printf("service: stopping")
			s <- svc.Status{State: svc.StopPending}
			cancel()
			app.Stop()
			log.Printf("service: stopped")
			return false, 0
		}
	}

	return false, 0
}

func isWindowsService() bool {
	ok, err := svc.IsWindowsService()
	return err == nil && ok
}

func installService() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("install: failed to get executable path: %v", err)
	}

	m, err := mgr.Connect()
	if err != nil {
		log.Fatalf("install: failed to connect to service manager: %v", err)
	}
	defer m.Disconnect()

	s, err := m.CreateService(serviceName, execPath, mgr.Config{
		StartType:   mgr.StartAutomatic,
		DisplayName: "Home Guard Agent",
		Description: "Agent de supervision et contrôle parental pour Home Assistant",
	})
	if err != nil {
		log.Fatalf("install: failed to create service: %v", err)
	}
	defer s.Close()

	log.Printf("service %q installed successfully", serviceName)
}

func uninstallService() {
	m, err := mgr.Connect()
	if err != nil {
		log.Fatalf("uninstall: failed to connect to service manager: %v", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		log.Fatalf("uninstall: service %q not found: %v", serviceName, err)
	}
	defer s.Close()

	if err := s.Delete(); err != nil {
		log.Fatalf("uninstall: failed to delete service: %v", err)
	}

	log.Printf("service %q uninstalled successfully", serviceName)
}

func runAsService() {
	execPath, _ := os.Executable()
	logPath := filepath.Join(filepath.Dir(execPath), "service.log")
	if logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	if err := svc.Run(serviceName, &windowsService{}); err != nil {
		log.Fatalf("service: run failed: %v", err)
	}
}
