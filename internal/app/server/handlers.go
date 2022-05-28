package server

import (
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ikashurnikov/shortener/internal/app/shortener"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	*chi.Mux
	shortener shortener.Shortener
	baseURL   url.URL
}

func NewHandler(shortener shortener.Shortener, baseURL url.URL) *Handler {
	badRequest := func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)
	}

	router := chi.NewRouter()
	router.Use(middleware.CleanPath)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.NotFound(badRequest)
	router.MethodNotAllowed(badRequest)

	handler := &Handler{
		Mux:       router,
		shortener: shortener,
		baseURL:   baseURL,
	}
	handler.Route("/", func(router chi.Router) {
		router.Post("/", handler.postLongLink())
		router.Get("/{shortURL}", handler.getShortLink())
	})
	return handler
}

func (handler *Handler) postLongLink() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			return
		}

		longURL, err := url.ParseRequestURI(string(bodyBytes))
		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusBadRequest)
			return
		}

		path, err := handler.shortener.Encode(longURL.String())
		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusBadRequest)
			return
		}

		shortURL := url.URL{
			Scheme: handler.baseURL.Scheme,
			Host:   handler.baseURL.Host,
			Path:   path,
		}

		responseWriter.WriteHeader(http.StatusCreated)
		responseWriter.Write([]byte(shortURL.String()))
	}
}

func (handler *Handler) getShortLink() http.HandlerFunc {
	return func(responseWriter http.ResponseWriter, request *http.Request) {
		shortURL := chi.URLParam(request, "shortURL")

		longURL, err := handler.shortener.Decode(shortURL)
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
