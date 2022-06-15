package main

import (
	"flag"
	"net/url"
	"os"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	SrvAddr         string  `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         url.URL `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string  `env:"FILE_STORAGE_PATH"`
}

func LoadConfig() (Config, error) {
	cfg := Config{}
	if err := cfg.parse(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (cfg *Config) parse() error {
	if err := env.Parse(cfg); err != nil {
		return err
	}

	flag.StringVar(&cfg.SrvAddr, "a", cfg.SrvAddr, "server address")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")

	flag.Func("b", cfg.BaseURL.String(), func(flagValue string) error {
		url, err := url.ParseRequestURI(flagValue)
		if err != nil {
			return err
		}
		cfg.BaseURL = *url
		return nil
	})
	return flag.CommandLine.Parse(os.Args[1:])
}
