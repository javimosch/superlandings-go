package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DatabasePath string
	DataDir      string
	LandingsDir  string
	SitesDir     string
	PIDFile      string
	LogFile      string
	UIDir        string // For UI override mode
	ServerPort   int
	AuthToken   string // API authentication token
	// Sync target configuration
	SyncTargetHost string
	SyncTargetUser string
	SyncTargetPort int
	SyncTargetKey  string
}

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	superlandingsDir := filepath.Join(homeDir, ".superlandings")
	dataDir := filepath.Join(superlandingsDir, "data")
	landingsDir := filepath.Join(superlandingsDir, "landings")
	sitesDir := filepath.Join(superlandingsDir, "sites")

	// Try to load config file
	configFile := filepath.Join(superlandingsDir, "config.json")
	cfg := &Config{
		DatabasePath: filepath.Join(superlandingsDir, "db.sql"),
		DataDir:      dataDir,
		LandingsDir:  landingsDir,
		SitesDir:     sitesDir,
		PIDFile:      filepath.Join(superlandingsDir, "sl-cli.pid"),
		LogFile:      filepath.Join(superlandingsDir, "sl-cli.log"),
		UIDir:        os.Getenv("SUPERLANDINGS_UI_DIR"),
		ServerPort:   8080,
	}

	if data, err := os.ReadFile(configFile); err == nil {
		if err := json.Unmarshal(data, cfg); err == nil {
			// Config loaded successfully
			// Fill in defaults for missing fields
			if cfg.DatabasePath == "" {
				cfg.DatabasePath = filepath.Join(superlandingsDir, "db.sql")
			}
			if cfg.DataDir == "" {
				cfg.DataDir = dataDir
			}
			if cfg.LandingsDir == "" {
				cfg.LandingsDir = landingsDir
			}
			if cfg.SitesDir == "" {
				cfg.SitesDir = sitesDir
			}
			if cfg.PIDFile == "" {
				cfg.PIDFile = filepath.Join(superlandingsDir, "sl-cli.pid")
			}
			if cfg.LogFile == "" {
				cfg.LogFile = filepath.Join(superlandingsDir, "sl-cli.log")
			}
			if cfg.ServerPort == 0 {
				cfg.ServerPort = 8080
			}
		}
		// If config file exists but can't be parsed, use defaults
	}

	return cfg, nil
}

func (c *Config) EnsureDirectories() error {
	dirs := []string{
		filepath.Dir(c.DatabasePath),
		c.DataDir,
		c.LandingsDir,
		c.SitesDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}