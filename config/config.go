package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Token string `yaml:"token"`
}

const configPath = "config.yml"

func New() (*Config, error) {
	cfgFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(cfgFile, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
