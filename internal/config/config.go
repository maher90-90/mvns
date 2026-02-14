package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Lang  string `json:"lang"`
	Theme string `json:"theme"`
}

func Default() Config {
	return Config{
		Lang:  "en",
		Theme: "dark",
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := Default()
		if err := save(path, &cfg); err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	if err != nil {
		return nil, err
	}

	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func ConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.Getenv("HOME")
	}
	return filepath.Join(dir, "mvns")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}
