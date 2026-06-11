package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

const dotDir = ".syncraft"

// WorldConfig is stored inside <world>/.syncraft/config.json.
type WorldConfig struct {
	Name   string `json:"name"`
	Server string `json:"server,omitempty"`
}

// Index is the global registry stored in the OS app-data dir.
// It maps world name → absolute path to the world save folder.
type Index struct {
	Worlds map[string]string `json:"worlds"`
}

// DefaultSavesDir returns the default Minecraft saves directory for the current OS.
func DefaultSavesDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appdata := os.Getenv("APPDATA")
		return filepath.Join(appdata, ".minecraft", "saves"), nil
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

// WorldConfigPath returns the path to the config file inside a world directory.
func WorldConfigPath(worldDir string) string {
	return filepath.Join(worldDir, dotDir, "config.json")
}

// SnapshotDir returns the path where the snapshot for a world is stored.
func SnapshotDir(name string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "syncraft", "worlds", name, "snapshot"), nil
}

// LoadWorldConfig reads the .syncraft/config.json from a world directory.
func LoadWorldConfig(worldDir string) (*WorldConfig, error) {
	data, err := os.ReadFile(WorldConfigPath(worldDir))
	if err != nil {
		return nil, err
	}
	var cfg WorldConfig
	return &cfg, json.Unmarshal(data, &cfg)
}

// Save writes the WorldConfig into <worldDir>/.syncraft/config.json.
func (c *WorldConfig) Save(worldDir string) error {
	dir := filepath.Join(worldDir, dotDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(WorldConfigPath(worldDir), data, 0644)
}

// indexPath returns the path to the global index file.
func indexPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "syncraft", "index.json"), nil
}

// LoadIndex reads the global world index, returning an empty one if it doesn't exist.
func LoadIndex() (*Index, error) {
	path, err := indexPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Index{Worlds: make(map[string]string)}, nil
	}
	if err != nil {
		return nil, err
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	if idx.Worlds == nil {
		idx.Worlds = make(map[string]string)
	}
	return &idx, nil
}

// Save writes the index to disk.
func (idx *Index) Save() error {
	path, err := indexPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
