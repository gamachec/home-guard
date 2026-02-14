package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Broker   string `json:"broker"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	ClientID string `json:"client_id"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
