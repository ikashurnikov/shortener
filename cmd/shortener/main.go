package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/ikashurnikov/shortener/internal/app/handler"
	"github.com/ikashurnikov/shortener/internal/app/storage"
	"github.com/ikashurnikov/shortener/internal/app/str2int"
	"github.com/ikashurnikov/shortener/internal/app/urlshortener"
)

const (
	host = "localhost"
	port = 8080
)

func main() {
	baseURL := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%v:%v", host, port),
	}

	server := http.Server{}
	server.Addr = fmt.Sprintf(":%d", port)

	shortener := urlshortener.StdShortener{
		Storage: storage.NewInMemoryStorage(),
		Encoder: str2int.NewZBase32Encoder(),
	}
	server.Handler = handler.NewHandler(&shortener, baseURL)
	log.Fatal(server.ListenAndServe())
}
