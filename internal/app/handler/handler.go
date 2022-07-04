package handler

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/asaskevich/govalidator"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ikashurnikov/shortener/internal/app/model"
	"github.com/ikashurnikov/shortener/internal/app/storage"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	*chi.Mux
	model     model.Model
	CipherKey string
}

func NewHandler(model model.Model, cipherKey string) *Handler {
	router := chi.NewRouter()

	handler := &Handler{
		Mux:       router,
		model:     model,
		CipherKey: cipherKey,
	}

	compressor := middleware.NewCompressor(flate.BestCompression)

	router.Use(middleware.CleanPath)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(decompressHandler)
	router.Use(compressor.Handler)

	handler.Route("/", func(router chi.Router) {
		router.Post("/", handler.postLongLink)
		router.Post("/api/shorten", handler.postAPIShorten)
		router.Post("/api/shorten/batch", handler.postAPIShortenBatch)
		router.Get("/api/user/urls", handler.getUserURLs)
		router.Get("/{shortURL}", handler.getShortLink)
		router.Get("/ping", handler.ping)
	})

	return handler
}

// POST /
func (h *Handler) postLongLink(rw http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	longURL := string(body)
	shortURL, err := h.shorten(req, rw, longURL)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.WriteHeader(http.StatusCreated)
	rw.Write([]byte(shortURL))
}

// GET /{shortURL}
func (h *Handler) getShortLink(rw http.ResponseWriter, req *http.Request) {
	shortLink := chi.URLParam(req, "shortURL")
	longLink, err := h.model.ExpandLink(shortLink)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(rw, req, longLink, http.StatusTemporaryRedirect)
}

// POST /api/shorten
func (h *Handler) postAPIShorten(rw http.ResponseWriter, req *http.Request) {
	type (
		Request struct {
			URL string `json:"url" valid:"required"`
		}

		Response struct {
			Result string `json:"result"`
		}
	)

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

	shortLink, err := h.shorten(req, rw, request.URL)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(rw)
	enc.SetEscapeHTML(false)
	err = enc.Encode(Response{Result: shortLink})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// /api/shorten/batch
func (h *Handler) postAPIShortenBatch(rw http.ResponseWriter, req *http.Request) {
	type (
		Request struct {
			CorrelationID string `json:"correlation_id"`
			OriginalURL   string `json:"original_url"`
		}

		Response struct {
			CorrelationID string `json:"correlation_id"`
			ShortURL      string `json:"short_url"`
		}
	)

	var request []Request
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	links := make([]string, len(request))
	for i, req := range request {
		links[i] = req.OriginalURL
	}

	shortLinks, err := h.shortenBatch(req, rw, links)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if len(shortLinks) != len(links) {
		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}

	response := make([]Response, len(shortLinks))
	for i := 0; i < len(response); i++ {
		response[i].CorrelationID = request[i].CorrelationID
		response[i].ShortURL = shortLinks[i]
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(rw)
	enc.SetEscapeHTML(false)
	err = enc.Encode(response)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GET /api/user/urls
func (h *Handler) getUserURLs(rw http.ResponseWriter, req *http.Request) {
	uid := h.getUserID(req)
	links, err := h.model.UserLinks(uid)

	if err != nil && !errors.Is(err, storage.ErrUserNotFound) {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if len(links) == 0 || (err != nil && errors.Is(err, storage.ErrUserNotFound)) {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	type Response struct {
		LongURL  string `json:"original_url"`
		ShortURL string `json:"short_url"`
	}

	response := make([]Response, 0, len(links))
	for l, s := range links {
		r := Response{
			LongURL:  l,
			ShortURL: s,
		}
		response = append(response, r)
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	enc := json.NewEncoder(rw)
	enc.SetEscapeHTML(false)
	err = enc.Encode(response)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GET /ping
func (h *Handler) ping(rw http.ResponseWriter, req *http.Request) {
	if err := h.model.Ping(); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

// shorten Cократить ссылку.
// Ф-ция так же укстанавлиает cookie "user_id".
func (h *Handler) shorten(req *http.Request, rw http.ResponseWriter, longLink string) (string, error) {
	userID := h.getUserID(req)

	shortLink, err := h.model.ShortenLink(&userID, longLink)
	if err != nil {
		return "", err
	}

	h.setUserID(rw, userID)
	return shortLink, nil
}

func (h *Handler) shortenBatch(req *http.Request, rw http.ResponseWriter, longLinks []string) ([]string, error) {
	userID := h.getUserID(req)

	shortLinks, err := h.model.ShortenLinks(&userID, longLinks)
	if err != nil {
		return nil, err
	}

	h.setUserID(rw, userID)
	return shortLinks, nil
}

func decompressHandler(next http.Handler) http.Handler {
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

func (h *Handler) getUserID(req *http.Request) model.UserID {
	cookie := NewSignedCookie(h.CipherKey)
	id, ok := cookie.GetInt(req, "user_id")
	if !ok {
		return model.InvalidUserID
	}
	return model.UserID(id)
}

func (h *Handler) setUserID(rw http.ResponseWriter, id model.UserID) {
	if id != model.InvalidUserID {
		NewSignedCookie(h.CipherKey).SetInt(rw, "user_id", int(id))
	}
}
