package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	ProjectAPIKey string `yaml:"project_api_key"`
	InstanceURL   string `yaml:"instance_url"`
	PollInterval  int    `yaml:"poll_interval"` // seconds
	Debug         bool   `yaml:"-"`              // runtime only, not saved to file
}

const (
	configFileName  = "ph-tui.yaml"
	defaultPollTime = 2 // seconds
)

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, configFileName), nil
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found. Run 'lazyhog login' to set up authentication")
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.InstanceURL == "" {
		cfg.InstanceURL = "https://app.posthog.com"
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = defaultPollTime
	}

	return &cfg, nil
}

// Save writes the configuration to disk
func Save(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Set defaults
	if cfg.InstanceURL == "" {
		cfg.InstanceURL = "https://app.posthog.com"
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = defaultPollTime
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with restricted permissions (0600 = owner read/write only)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Exists checks if the configuration file exists
func Exists() bool {
	path, err := GetConfigPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}
