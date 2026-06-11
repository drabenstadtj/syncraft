package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type World struct {
	Path     string `json:"path"`
	Snapshot string `json:"snapshot"`
}

type Config struct {
	Worlds map[string]World `json:"worlds"`
}

// configPath returns the path to the syncraft config file.
func configPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "syncraft", "config.json"), nil
}

// SnapshotDir returns the snapshot directory for a world name.
func SnapshotDir(name string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "syncraft", "worlds", name, "snapshot"), nil
}

// Load reads the config from disk. Returns an empty config if it doesn't exist yet.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{Worlds: make(map[string]World)}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Worlds == nil {
		cfg.Worlds = make(map[string]World)
	}
	return &cfg, nil
}

// Save writes the config to disk, creating directories as needed.
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
