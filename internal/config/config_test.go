package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	content := `{
		"broker": "localhost",
		"port": 1883,
		"username": "user",
		"password": "pass",
		"client_id": "test-pc"
	}`

	f, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Broker != "localhost" {
		t.Errorf("Broker = %q, want %q", cfg.Broker, "localhost")
	}
	if cfg.Port != 1883 {
		t.Errorf("Port = %d, want %d", cfg.Port, 1883)
	}
	if cfg.Username != "user" {
		t.Errorf("Username = %q, want %q", cfg.Username, "user")
	}
	if cfg.Password != "pass" {
		t.Errorf("Password = %q, want %q", cfg.Password, "pass")
	}
	if cfg.ClientID != "test-pc" {
		t.Errorf("ClientID = %q, want %q", cfg.ClientID, "test-pc")
	}
}
