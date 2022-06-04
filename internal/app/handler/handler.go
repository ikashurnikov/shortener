package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/ikashurnikov/shortener/internal/app/urlshortener"
)

type Handler struct {
	*chi.Mux
	urlShortener urlshortener.Shortener
	baseURL      url.URL
}

func NewHandler(urlShortener urlshortener.Shortener, baseURL url.URL) *Handler {
	router := chi.NewRouter()
	router.Use(middleware.CleanPath)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)

	handler := &Handler{
		Mux:          router,
		urlShortener: urlShortener,
		baseURL:      baseURL,
	}
	handler.Route("/", func(router chi.Router) {
		router.Post("/", handler.postLongLink)
		router.Get("/{shortURL}", handler.getShortLink)
	})
	return handler
}

func (handler *Handler) postLongLink(rw http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return
	}

	longURL := string(body)
	path, err := handler.urlShortener.EncodeLongURL(longURL)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL := url.URL{
		Scheme: handler.baseURL.Scheme,
		Host:   handler.baseURL.Host,
		Path:   path,
	}

	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte(shortURL.String()))
}

func (handler *Handler) getShortLink(rw http.ResponseWriter, req *http.Request) {
	shortURL := chi.URLParam(req, "shortURL")

	longURL, err := handler.urlShortener.DecodeShortURL(shortURL)
	if err != nil {
		http.Error(
			rw,
			fmt.Sprintf("Failed to decode URL %v : %v", shortURL, err.Error()),
			http.StatusBadRequest,
		)
		return
	}

	http.Redirect(rw, req, longURL, http.StatusTemporaryRedirect)
}
