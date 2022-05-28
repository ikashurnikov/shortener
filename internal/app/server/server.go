package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/ikashurnikov/shortener/internal/app/shortener"
)

type Server struct {
	shortener shortener.Shortener
	http.Server
}

func NewServer(shortener shortener.Shortener) *Server {
	return &Server{
		shortener: shortener,
	}
}

func (server *Server) Run(host string, port uint16) {
	baseURL := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%v:%v", host, port),
	}

	server.Addr = fmt.Sprintf(":%d", port)
	server.Handler = NewHandler(server.shortener, baseURL)
	log.Fatal(server.ListenAndServe())
}
