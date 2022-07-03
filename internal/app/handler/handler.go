package handler

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/asaskevich/govalidator"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ikashurnikov/shortener/internal/app/storage"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Handler struct {
	*chi.Mux
	storage   storage.Storage
	baseURL   url.URL
	CipherKey string
}

func NewHandler(storage storage.Storage, baseURL url.URL, cipherKey string) *Handler {
	router := chi.NewRouter()

	handler := &Handler{
		Mux:       router,
		storage:   storage,
		baseURL:   baseURL,
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
	shortURL := chi.URLParam(req, "shortURL")
	longURL, err := h.storage.GetLongURL(shortURL)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(rw, req, longURL, http.StatusTemporaryRedirect)
}

// POST /api/shorten
func (h *Handler) postAPIShorten(rw http.ResponseWriter, req *http.Request) {
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

	shortURL, err := h.shorten(req, rw, request.URL)
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

// GET /api/user/urls
func (h *Handler) getUserURLs(rw http.ResponseWriter, req *http.Request) {
	uid := h.getUserID(req)
	urls, err := h.storage.GetUserURLs(uid, h.baseURL)

	if err != nil && !errors.Is(err, storage.ErrUserNotFound) {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if len(urls) == 0 || (err != nil && errors.Is(err, storage.ErrUserNotFound)) {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	enc := json.NewEncoder(rw)
	enc.SetEscapeHTML(false)
	err = enc.Encode(urls)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GET /ping
func (h *Handler) ping(rw http.ResponseWriter, req *http.Request) {
	if err := h.storage.Ping(); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}

// shorten Cократить ссылку.
// Ф-ция так же укстанавлиает cookie "user_id".
func (h *Handler) shorten(req *http.Request, rw http.ResponseWriter, longURL string) (string, error) {
	userID := h.getUserID(req)

	shortURL, err := h.storage.AddLongURL(&userID, longURL, h.baseURL)
	if err != nil {
		return "", err
	}

	h.setUserID(rw, userID)
	return shortURL, nil
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

func (h *Handler) getUserID(req *http.Request) storage.UserID {
	cookie := NewSignedCookie(h.CipherKey)
	id, ok := cookie.GetInt(req, "user_id")
	if !ok {
		return storage.InvalidUserID
	}
	return storage.UserID(id)
}

func (h *Handler) setUserID(rw http.ResponseWriter, id storage.UserID) {
	if id != storage.InvalidUserID {
		NewSignedCookie(h.CipherKey).SetInt(rw, "user_id", int(id))
	}
}
