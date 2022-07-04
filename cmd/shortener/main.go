package main

import (
	"github.com/ikashurnikov/shortener/internal/app/model"
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
	m := model.New(repo, cfg.BaseURL)
	server.Handler = handler.NewHandler(m, "secret")
	log.Fatal(server.ListenAndServe())
}

func newStorage(cfg *Config) storage.Storage {
	switch {
	case cfg.DatabaseDSN != "":
		db, err := storage.NewDBStorage(cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		return db

	case cfg.FileStoragePath != "":
		fileStorage, err := storage.NewFileStorage(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		return fileStorage
	}

	return storage.NewInMemoryStorage()
}
