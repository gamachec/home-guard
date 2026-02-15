//go:build windows

package main

import (
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "HomeGuard"

func installService() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("install: failed to get executable path: %v", err)
	}

	updaterPath := filepath.Join(filepath.Dir(execPath), "home-guard-updater.exe")

	m, err := mgr.Connect()
	if err != nil {
		log.Fatalf("install: failed to connect to service manager: %v", err)
	}
	defer m.Disconnect()

	s, err := m.CreateService(serviceName, updaterPath, mgr.Config{
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
