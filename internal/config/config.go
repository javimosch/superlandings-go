package config

import (
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

	return &Config{
		DatabasePath: filepath.Join(superlandingsDir, "db.sql"),
		DataDir:      dataDir,
		LandingsDir:  landingsDir,
		SitesDir:     sitesDir,
		PIDFile:      filepath.Join(superlandingsDir, "sl-cli.pid"),
		LogFile:      filepath.Join(superlandingsDir, "sl-cli.log"),
		UIDir:        os.Getenv("SUPERLANDINGS_UI_DIR"),
		ServerPort:   8080,
	}, nil
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