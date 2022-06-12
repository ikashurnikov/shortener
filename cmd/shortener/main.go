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

type Config struct {
	SrvAddr string  `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL url.URL `env:"BASE_URL" envDefault:"http://localhost:8080"`
}

func main() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	server := http.Server{}
	server.Addr = cfg.SrvAddr

	shortener := urlshortener.StdShortener{
		Storage: storage.NewInMemoryStorage(),
		Encoder: str2int.NewZBase32Encoder(),
	}
	server.Handler = handler.NewHandler(&shortener, cfg.BaseURL)
	log.Fatal(server.ListenAndServe())
}
