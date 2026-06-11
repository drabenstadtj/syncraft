package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

const DotDir = ".syncraft"

type WorldConfig struct {
	Server string `json:"server,omitempty"`
}

// WorldConfigPath returns the path to config.json inside a world directory.
func WorldConfigPath(worldDir string) string {
	return filepath.Join(worldDir, DotDir, "config.json")
}

// SnapshotDir returns the snapshot directory inside a world directory.
func SnapshotDir(worldDir string) string {
	return filepath.Join(worldDir, DotDir, "snapshot")
}

// LoadWorldConfig reads .syncraft/config.json from a world directory.
func LoadWorldConfig(worldDir string) (*WorldConfig, error) {
	data, err := os.ReadFile(WorldConfigPath(worldDir))
	if err != nil {
		return nil, err
	}
	var cfg WorldConfig
	return &cfg, json.Unmarshal(data, &cfg)
}

// Save writes WorldConfig into <worldDir>/.syncraft/config.json.
func (c *WorldConfig) Save(worldDir string) error {
	dir := filepath.Join(worldDir, DotDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(WorldConfigPath(worldDir), data, 0644)
}

// DefaultSavesDir returns the default Minecraft saves directory for the current OS.
func DefaultSavesDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), ".minecraft", "saves"), nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "minecraft", "saves"), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".minecraft", "saves"), nil
	}
}

// IsInitialized reports whether a world directory has a .syncraft directory.
func IsInitialized(worldDir string) bool {
	_, err := os.Stat(filepath.Join(worldDir, DotDir))
	return err == nil
}
