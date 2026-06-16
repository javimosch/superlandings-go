package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var cliConfigMu sync.Mutex

const (
	cliConfigFile = "cli_config.json"
)

type CLIConfig struct {
	Targets []Target `json:"targets"`
}

type Target struct {
	Name      string `json:"name"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	AuthToken string `json:"auth_token,omitempty"`
	Default   bool   `json:"default"`
}

func LoadCLIConfig() (*CLIConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	
	configPath := filepath.Join(homeDir, ".superlandings", cliConfigFile)
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &CLIConfig{Targets: []Target{}}, nil
		}
		return nil, err
	}
	
	var config CLIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}

func SaveCLIConfig(config *CLIConfig) error {
	cliConfigMu.Lock()
	defer cliConfigMu.Unlock()
	
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	
	configPath := filepath.Join(homeDir, ".superlandings", cliConfigFile)
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(configPath, data, 0600)
}

func GetDefaultTarget() (*Target, error) {
	config, err := LoadCLIConfig()
	if err != nil {
		return nil, err
	}
	
	for _, target := range config.Targets {
		if target.Default {
			return &target, nil
		}
	}
	
	return nil, fmt.Errorf("no default target configured")
}

func GetTarget(name string) (*Target, error) {
	config, err := LoadCLIConfig()
	if err != nil {
		return nil, err
	}
	
	for _, target := range config.Targets {
		if target.Name == name {
			return &target, nil
		}
	}
	
	return nil, fmt.Errorf("target '%s' not found", name)
}

func AddTarget(target Target) error {
	config, err := LoadCLIConfig()
	if err != nil {
		return err
	}
	
	// Check if target already exists
	for i, t := range config.Targets {
		if t.Name == target.Name {
			// Update existing target
			config.Targets[i] = target
			return SaveCLIConfig(config)
		}
	}
	
	// Add new target
	config.Targets = append(config.Targets, target)
	return SaveCLIConfig(config)
}

func RemoveTarget(name string) error {
	config, err := LoadCLIConfig()
	if err != nil {
		return err
	}
	
	for i, target := range config.Targets {
		if target.Name == name {
			config.Targets = append(config.Targets[:i], config.Targets[i+1:]...)
			return SaveCLIConfig(config)
		}
	}
	
	return fmt.Errorf("target '%s' not found", name)
}