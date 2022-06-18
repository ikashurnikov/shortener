package handler

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ikashurnikov/shortener/internal/app/urlshortener"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Handler struct {
	*chi.Mux
	shortener urlshortener.Shortener
	baseURL   url.URL
}

func NewHandler(urlShortener urlshortener.Shortener, baseURL url.URL) *Handler {
	router := chi.NewRouter()
	router.Use(middleware.CleanPath)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(decompressor)
	compressor := middleware.NewCompressor(flate.BestCompression)
	router.Use(compressor.Handler)

	handler := &Handler{
		Mux:       router,
		shortener: urlShortener,
		baseURL:   baseURL,
	}

	handler.Route("/", func(router chi.Router) {
		router.Post("/", handler.postLongLink)
		router.Get("/{shortURL}", handler.getShortLink)
		router.Post("/api/shorten", handler.postAPIShorten)
	})
	return handler
}

// POST /
func (handler *Handler) postLongLink(rw http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return
	}

	longURL := string(body)
	shortURL, err := handler.shorten(longURL)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte(shortURL))
}

// GET /{shortURL}
func (handler *Handler) getShortLink(rw http.ResponseWriter, req *http.Request) {
	shortURL := chi.URLParam(req, "shortURL")

	longURL, err := handler.shortener.DecodeShortURL(shortURL)
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

// POST /api/shorten
func (handler *Handler) postAPIShorten(rw http.ResponseWriter, req *http.Request) {
	type Request struct {
		URL string `json:"url" valid:"required"`
	}
	type Response struct {
		Result string `json:"result"`
	}

	var request Request
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	valid, err := govalidator.ValidateStruct(request)
	if err != nil || !valid {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, err := handler.shorten(request.URL)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(rw)
	enc.SetEscapeHTML(false)
	err = enc.Encode(Response{Result: shortURL})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (handler *Handler) shorten(longURL string) (string, error) {
	shortPath, err := handler.shortener.EncodeLongURL(longURL)
	if err != nil {
		return "", err
	}

	shortURL := url.URL{
		Scheme: handler.baseURL.Scheme,
		Host:   handler.baseURL.Host,
		Path:   shortPath,
	}

	return shortURL.String(), nil
}

func decompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var bodyDecompressor io.ReadCloser

		encoding := strings.ToLower(req.Header.Get("Content-Encoding"))
		switch encoding {
		case "gzip":
			gz, err := gzip.NewReader(req.Body)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}
			bodyDecompressor = gz
		case "deflate":
			bodyDecompressor = flate.NewReader(req.Body)
		}

		if bodyDecompressor == nil {
			next.ServeHTTP(rw, req)
			return
		}

		req.Body = bodyDecompressor
		defer func() {
			bodyDecompressor.Close()
		}()

		next.ServeHTTP(rw, req)
	})
}
