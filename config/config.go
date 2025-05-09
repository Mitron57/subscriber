package config

import (
	"errors"
	"os"
)

type Config struct {
	Host string
	Port string
}

func NewConfig() (*Config, error) {
	var config Config
	config.Host = os.Getenv("SUBSCRIBER_HOST")
	if config.Host == "" {
		return nil, errors.New("HOST environment variable not set")
	}
	config.Port = os.Getenv("SUBSCRIBER_PORT")
	if config.Port == "" {
		return nil, errors.New("PORT environment variable not set")
	}
	return &config, nil
}
