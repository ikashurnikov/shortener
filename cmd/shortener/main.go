package main

import (
	"github.com/ikashurnikov/shortener/internal/app/server"
	"github.com/ikashurnikov/shortener/internal/app/shortener"
)

func main() {
	shortener := shortener.NewZBase32Shortener()
	srv := server.NewServer(&shortener)
	srv.ListenAndServe(8080)
}
