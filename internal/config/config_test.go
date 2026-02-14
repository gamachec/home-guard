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

func TestLoadBlacklist(t *testing.T) {
	content := `{"broker":"localhost","port":1883,"client_id":"pc","blacklist":["game.exe","browser.exe"]}`

	f, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Blacklist) != 2 {
		t.Fatalf("Blacklist length = %d, want 2", len(cfg.Blacklist))
	}
	if cfg.Blacklist[0] != "game.exe" {
		t.Errorf("Blacklist[0] = %q, want %q", cfg.Blacklist[0], "game.exe")
	}
}

func TestSave(t *testing.T) {
	f, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	cfg := &Config{
		Broker:    "localhost",
		Port:      1883,
		ClientID:  "pc",
		Blacklist: []string{"game.exe"},
	}

	if err := Save(f.Name(), cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(f.Name())
	if err != nil {
		t.Fatalf("Load() after Save() error = %v", err)
	}

	if loaded.Broker != cfg.Broker {
		t.Errorf("Broker = %q, want %q", loaded.Broker, cfg.Broker)
	}
	if len(loaded.Blacklist) != 1 || loaded.Blacklist[0] != "game.exe" {
		t.Errorf("Blacklist = %v, want [game.exe]", loaded.Blacklist)
	}
}
