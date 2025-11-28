package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Upstream UpstreamConfig `yaml:"upstream"`
	LinuxDo  LinuxDoConfig  `yaml:"linuxdo"`
	JWT      JWTConfig      `yaml:"jwt"`
}

type LinuxDoConfig struct {
	ClientID       string `yaml:"client_id"`
	ClientSecret   string `yaml:"client_secret"`
	BackendBaseURL string `yaml:"backend_base_url"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type UpstreamConfig struct {
	BaseURL string `yaml:"base_url"`
}

var GlobalConfig *Config

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	GlobalConfig = &cfg
	return &cfg, nil
}
