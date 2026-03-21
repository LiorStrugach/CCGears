package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultDirName = ".ccgears"
	configFileName = "config.json"
	presetsDirName = "presets"
	backupDirName  = "backup"
	backupSlotName = "last"
)

// Config holds CCGears configuration and resolved paths.
type Config struct {
	Home          string `json:"-"`
	StorePath     string `json:"-"`
	BackupPath    string `json:"-"`
	DefaultPreset string `json:"default_preset,omitempty"`
	ActivePreset  string `json:"active_preset,omitempty"`
	StorePathCfg  string `json:"store_path,omitempty"`
}

// Load resolves the CCGears home directory and loads config.json.
// Priority: CCGEARS_HOME env > config.json store_path > ~/.ccgears default.
func Load() (*Config, error) {
	home, err := resolveHome()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Home:      home,
		StorePath: filepath.Join(home, presetsDirName),
	}
	cfg.BackupPath = filepath.Join(home, backupDirName, backupSlotName)

	// Try to load existing config.json
	cfgFile := filepath.Join(home, configFileName)
	data, err := os.ReadFile(cfgFile)
	if err == nil {
		_ = json.Unmarshal(data, cfg)
		if cfg.StorePathCfg != "" {
			cfg.StorePath = cfg.StorePathCfg
		}
	}

	return cfg, nil
}

// EnsureDirs creates the CCGears directory tree if it doesn't exist.
func (c *Config) EnsureDirs() error {
	dirs := []string{c.Home, c.StorePath, filepath.Join(c.Home, backupDirName)}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("creating %s: %w", d, err)
		}
	}
	return nil
}

// Save writes config.json to the CCGears home directory.
func (c *Config) Save() error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	cfgFile := filepath.Join(c.Home, configFileName)
	tmp := cfgFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, cfgFile)
}

func resolveHome() (string, error) {
	// Check env var first
	if env := os.Getenv("CCGEARS_HOME"); env != "" {
		return env, nil
	}

	// Default: ~/.ccgears
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, defaultDirName), nil
}
