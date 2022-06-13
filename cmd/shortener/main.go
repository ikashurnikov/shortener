package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/caarlos0/env/v6"

	"github.com/ikashurnikov/shortener/internal/app/handler"
	"github.com/ikashurnikov/shortener/internal/app/storage"
	"github.com/ikashurnikov/shortener/internal/app/str2int"
	"github.com/ikashurnikov/shortener/internal/app/urlshortener"
)

type config struct {
	SrvAddr         string  `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         url.URL `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string  `env:"FILE_STORAGE_PATH"`
}

func main() {
	cfg := config{}
	cfg.parse()

	repo := newStorage(&cfg)
	defer repo.Close()

	server := http.Server{
		Addr: cfg.SrvAddr,
	}

	shortener := urlshortener.StdShortener{
		Storage: repo,
		Encoder: str2int.NewZBase32Encoder(),
	}
	server.Handler = handler.NewHandler(&shortener, cfg.BaseURL)
	log.Fatal(server.ListenAndServe())
}

func (cfg *config) parse() {
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}
}

func newStorage(cfg *config) storage.Storage {
	if cfg.FileStoragePath != "" {
		fileStorage, err := storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		return fileStorage
	}
	return storage.NewInMemoryStorage()
}
