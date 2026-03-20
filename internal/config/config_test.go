package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Default(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CCGEARS_HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Home != dir {
		t.Errorf("Home = %q, want %q", cfg.Home, dir)
	}
	if cfg.StorePath != filepath.Join(dir, "presets") {
		t.Errorf("StorePath = %q, want %q", cfg.StorePath, filepath.Join(dir, "presets"))
	}
	if cfg.BackupPath != filepath.Join(dir, "backup", "last") {
		t.Errorf("BackupPath = %q, want %q", cfg.BackupPath, filepath.Join(dir, "backup", "last"))
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CCGEARS_HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Home != dir {
		t.Errorf("Home = %q, want %q", cfg.Home, dir)
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CCGEARS_HOME", dir)

	// Write a config file with custom store path
	customStore := filepath.Join(dir, "custom-presets")
	cfgData, _ := json.Marshal(map[string]string{
		"store_path":     customStore,
		"default_preset": "my-default",
	})
	os.WriteFile(filepath.Join(dir, "config.json"), cfgData, 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.StorePath != customStore {
		t.Errorf("StorePath = %q, want %q", cfg.StorePath, customStore)
	}
	if cfg.DefaultPreset != "my-default" {
		t.Errorf("DefaultPreset = %q, want %q", cfg.DefaultPreset, "my-default")
	}
}

func TestEnsureDirs(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CCGEARS_HOME", dir)

	cfg, _ := Load()
	if err := cfg.EnsureDirs(); err != nil {
		t.Fatal(err)
	}

	for _, d := range []string{cfg.Home, cfg.StorePath, filepath.Join(cfg.Home, "backup")} {
		info, err := os.Stat(d)
		if err != nil {
			t.Errorf("dir %q not created: %v", d, err)
		} else if !info.IsDir() {
			t.Errorf("%q is not a directory", d)
		}
	}
}

func TestSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CCGEARS_HOME", dir)

	cfg, _ := Load()
	cfg.DefaultPreset = "test-preset"
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}

	cfg2, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.DefaultPreset != "test-preset" {
		t.Errorf("DefaultPreset = %q, want %q", cfg2.DefaultPreset, "test-preset")
	}
}
