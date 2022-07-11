package main

import (
	"github.com/ikashurnikov/shortener/internal/app/repo"
	"github.com/ikashurnikov/shortener/internal/app/service"
	"log"
	"net/http"

	"github.com/ikashurnikov/shortener/internal/app/handler"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	repo := newRepo(&cfg)
	defer repo.Close()

	server := http.Server{
		Addr: cfg.SrvAddr,
	}
	m := service.NewShortener(repo, cfg.BaseURL)
	server.Handler = handler.NewHandler(m, "secret")
	log.Fatal(server.ListenAndServe())
}

func newRepo(cfg *Config) repo.Repo {
	switch {
	case cfg.DatabaseDSN != "":
		db, err := repo.NewDBRepo(cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		return db

	case cfg.FileStoragePath != "":
		fileStorage, err := repo.NewFileRepo(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		return fileStorage
	}

	return repo.NewInMemoryRepo()
}
