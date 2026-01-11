package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

const ConfigFileName = "config.json"

type Config struct {
	Token string `json:"token"`
	Host  string `json:"host"`
	Port  string `json:"port"`
}

// LoadOrInit tries to load the config from disk.
// If it fails or file doesn't exist, it generates a new token and saves it.
func LoadOrInit() (*Config, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(filepath.Dir(exePath), ConfigFileName)

	// Try read
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err == nil && cfg.Token != "" {
			return &cfg, nil
		}
	}

	// Generate new
	cfg := &Config{
		Token: generateToken(),
		Host:  "0.0.0.0",
		Port:  "5000",
	}

	// Save
	data, _ := json.MarshalIndent(cfg, "", "  ")
	_ = os.WriteFile(configPath, data, 0644)

	return cfg, nil
}

func generateToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
