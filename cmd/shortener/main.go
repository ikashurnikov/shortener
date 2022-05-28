package main

import (
	"github.com/ikashurnikov/shortener/internal/app/server"
	"github.com/ikashurnikov/shortener/internal/app/shortener"
	"github.com/ikashurnikov/shortener/internal/app/storage"
)

func main() {
	storage := storage.NewInMemoryStorage()
	shortener := shortener.NewZBase32Shortener(storage)
	srv := server.NewServer(&shortener)
	srv.ListenAndServe(8080)
}
