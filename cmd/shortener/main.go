package main

import (
	"log"
	"net/http"

	"github.com/ikashurnikov/shortener/internal/app/handler"
	"github.com/ikashurnikov/shortener/internal/app/storage"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	repo := newStorage(&cfg)
	defer repo.Close()

	server := http.Server{
		Addr: cfg.SrvAddr,
	}

	server.Handler = handler.NewHandler(repo, cfg.BaseURL, "secret")
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
