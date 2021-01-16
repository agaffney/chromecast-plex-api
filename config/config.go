package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Address string `default:"0.0.0.0"`
	Port    int    `default:"8080"`
}

var config *Config

func Get() *Config {
	return config
}

func Load() (*Config, error) {
	config = &Config{}
	err := envconfig.Process("", config)
	return config, err
}
