package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// Config file name
	ConfigFileName = "config.yaml"
	ConfigDirName  = ".ohmymem"
)

// Config represents user configuration
type Config struct {
	Init InitConfig `yaml:"init"`
}

// InitConfig holds init command defaults
type InitConfig struct {
	Yes bool `yaml:"yes"`
}

// Load loads configuration from config file
func Load() (*Config, error) {
	cfg := &Config{
		Init: InitConfig{
			Yes: false,
		},
	}

	// Load from config file (if exists)
	if err := cfg.loadFromFile(); err != nil {
		// Ignore file not found, but log other errors
		if !os.IsNotExist(err) {
			return cfg, err
		}
	}

	return cfg, nil
}

// GetConfigPath returns the config file path
func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ConfigDirName, ConfigFileName)
}

// loadFromFile loads config from YAML file
func (c *Config) loadFromFile() error {
	path := GetConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fileConfig Config
	if err := yaml.Unmarshal(data, &fileConfig); err != nil {
		return err
	}

	// Merge file config
	c.Init.Yes = fileConfig.Init.Yes

	return nil
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[1:])
	}
	return path
}
