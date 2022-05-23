package server

import (
	"fmt"
	"github.com/ikashurnikov/shortener/internal/app/shortener"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func route(addLongLink http.HandlerFunc, getShortLink http.HandlerFunc) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			addLongLink(responseWriter, request)
			return
		}

		if request.Method == http.MethodGet {
			getShortLink(responseWriter, request)
			return
		}

		http.Error(responseWriter, "Method not allowed", http.StatusBadRequest)
	}
}

func addLongLink(shortener shortener.Shortener, baseURL url.URL) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.Error(responseWriter, "Bad request", http.StatusBadRequest)
			return
		}

		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			return
		}

		longURL := string(bodyBytes)
		log.Printf("POST: Reviced url %s", longURL)

		path, err := shortener.Encode(longURL)
		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusBadRequest)
			return
		}

		shortURL := url.URL{
			Scheme: baseURL.Scheme,
			Host:   baseURL.Host,
			Path:   path,
		}

		responseWriter.WriteHeader(http.StatusCreated)
		responseWriter.Write([]byte(shortURL.String()))
	}
}

func getShortLink(shortener shortener.Shortener) http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		shortURL := strings.Trim(request.URL.Path, "/")
		if shortURL == "" {
			http.Error(responseWriter, "Empty URL", http.StatusBadRequest)
			return
		}

		log.Printf("Get: Reviced url %s", shortURL)

		longURL, err := shortener.Decode(shortURL)
		if err != nil {
			http.Error(
				responseWriter,
				fmt.Sprintf("Failed to decode URL %v : %v", shortURL, err.Error()),
				http.StatusBadRequest,
			)
			return
		}

		responseWriter.Header().Set("Location", longURL)
		responseWriter.WriteHeader(http.StatusTemporaryRedirect)
		responseWriter.Write(nil)
	}
}
