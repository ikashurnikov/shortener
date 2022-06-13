package main

import (
	"flag"
	"log"
	"net/url"

	"github.com/caarlos0/env/v6"
)

var (
	defaultServeAddress string  = ":8080"
	defaultBaseURL      url.URL = url.URL{Scheme: "http", Host: "localhost:8080"}
)

type Config struct {
	SrvAddr         string  `env:"SERVER_ADDRESS"`
	BaseURL         url.URL `env:"BASE_URL"`
	FileStoragePath string  `env:"FILE_STORAGE_PATH"`
}

func (cfg *Config) Parse() {
	cmdLineCfg := Config{}
	cmdLineCfg.parseCommandLine()

	cfg.parseEnv()

	if cfg.SrvAddr == "" {
		cfg.SrvAddr = cmdLineCfg.SrvAddr
	}

	if cfg.BaseURL.String() == "" {
		cfg.BaseURL = cmdLineCfg.BaseURL
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = cmdLineCfg.FileStoragePath
	}
}

func (cfg *Config) parseCommandLine() {
	flag.StringVar(&cfg.SrvAddr, "a", defaultServeAddress, "server address")
	flag.StringVar(&cfg.FileStoragePath, "f", "", "file storage path")

	cfg.BaseURL = defaultBaseURL
	flag.Func("b", defaultBaseURL.String(), func(flagValue string) error {
		url, err := url.ParseRequestURI(flagValue)
		if err != nil {
			return err
		}
		cfg.BaseURL = *url
		return nil
	})
	flag.Parse()
}

func (cfg *Config) parseEnv() {
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}
}
