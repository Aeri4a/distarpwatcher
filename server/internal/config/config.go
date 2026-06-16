package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Port      string `yaml:"port"`
	CaCert    string `yaml:"ca_cert"`
	ServerPem string `yaml:"server_pem"`
	ServerKey string `yaml:"server_key"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

type APIConfig struct {
	Port string `yaml:"port"`
}

type MailConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	From     string `yaml:"from"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	API      APIConfig      `yaml:"api"`
	Mail     MailConfig     `yaml:"mail"`
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		path = "config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	return &cfg, nil
}
