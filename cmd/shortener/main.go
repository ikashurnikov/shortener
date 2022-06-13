package main

import (
	"log"
	"net/http"

	"github.com/ikashurnikov/shortener/internal/app/handler"
	"github.com/ikashurnikov/shortener/internal/app/storage"
	"github.com/ikashurnikov/shortener/internal/app/str2int"
	"github.com/ikashurnikov/shortener/internal/app/urlshortener"
)

func main() {
	cfg := Config{}
	cfg.Parse()

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

func newStorage(cfg *Config) storage.Storage {
	if cfg.FileStoragePath != "" {
		fileStorage, err := storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		return fileStorage
	}
	return storage.NewInMemoryStorage()
}
