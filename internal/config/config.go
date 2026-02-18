package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// ServerConfig holds the configuration for a single TeamSpeak server
type ServerConfig struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Protocol string `yaml:"protocol"` // "tcp" or "ssh", default "tcp"
}

// Config holds the global configuration
type Config struct {
	Servers      []ServerConfig `yaml:"servers"`
	MetricsPort  int            `yaml:"metrics_port"`
	ReadInterval int            `yaml:"read_interval"`
}

// LoadConfig reads the configuration from the given file path
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults if not provided
	if config.MetricsPort == 0 {
		config.MetricsPort = 8000
	}
	if config.ReadInterval == 0 {
		config.ReadInterval = 60
	}

	return &config, nil
}
