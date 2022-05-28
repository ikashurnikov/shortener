package server

import (
	"fmt"
	"github.com/ikashurnikov/shortener/internal/app/shortener"
	"log"
	"net/http"
	"net/url"
)

type Server struct {
	shortener shortener.Shortener
}

func NewServer(shortener shortener.Shortener) Server {
	return Server{shortener: shortener}
}

func (server *Server) ListenAndServe(port uint16) {
	baseURL := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%v", port),
	}

	routingTable := routingTable{
		http.MethodGet:  getShortLink(server.shortener),
		http.MethodPost: addLongLink(server.shortener, baseURL),
	}

	http.HandleFunc("/", route(routingTable))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
