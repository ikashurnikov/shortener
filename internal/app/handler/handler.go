package handler

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ikashurnikov/shortener/internal/app/model"
	"github.com/ikashurnikov/shortener/internal/app/service"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	*chi.Mux
	shortener service.Shortener
	workers   errgroup.Group
	CipherKey string
}

const (
	workersCount = 10
)

func NewHandler(shortener service.Shortener, cipherKey string) *Handler {
	router := chi.NewRouter()

	handler := &Handler{
		Mux:       router,
		shortener: shortener,
		CipherKey: cipherKey,
		workers:   errgroup.Group{},
	}
	handler.workers.SetLimit(workersCount)

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
		router.Delete("/api/user/urls", handler.deleteURLs)
		router.Get("/ping", handler.ping)
	})

	return handler
}

func (h *Handler) Shutdown() {
	_ = h.workers.Wait()
}

// POST /
func (h *Handler) postLongLink(rw http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	link, err := h.shorten(req, rw, originalURL)

	status := http.StatusCreated

	if err != nil {
		if !errors.Is(err, model.ErrLinkAlreadyExists) {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		status = http.StatusConflict
	}

	rw.WriteHeader(status)
	rw.Write([]byte(link.ShortURL))
}

// GET /{shortURL}
func (h *Handler) getShortLink(rw http.ResponseWriter, req *http.Request) {
	shortURL := chi.URLParam(req, "shortURL")
	link, err := h.shortener.GetLinkByShortURL(shortURL)

	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, model.ErrLinkRemoved) {
			status = http.StatusGone
		}
		http.Error(rw, err.Error(), status)
		return
	}

	http.Redirect(rw, req, link.OriginalURL, http.StatusTemporaryRedirect)
}

// POST /api/shorten
func (h *Handler) postAPIShorten(rw http.ResponseWriter, req *http.Request) {
	type (
		Request struct {
			URL string `json:"url" valid:"required"`
		}

		Reply struct {
			Result string `json:"result"`
		}
	)

	var request Request

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	status := http.StatusCreated

	link, err := h.shorten(req, rw, request.URL)
	if err != nil {
		if !errors.Is(err, model.ErrLinkAlreadyExists) {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		status = http.StatusConflict
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)

	enc := json.NewEncoder(rw)
	enc.SetEscapeHTML(false)
	err = enc.Encode(Reply{Result: link.ShortURL})
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

		Reply struct {
			CorrelationID string `json:"correlation_id"`
			ShortURL      string `json:"short_url"`
		}
	)

	var request []Request
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	origURLs := make([]string, len(request))
	for i, req := range request {
		origURLs[i] = req.OriginalURL
	}

	links, err := h.shortenBatch(req, rw, origURLs)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if len(links) != len(origURLs) {
		http.Error(rw, "internal error", http.StatusInternalServerError)
		return
	}

	reply := make([]Reply, len(links))
	for i, link := range links {
		idx := slices.IndexFunc(request, func(req Request) bool {
			return link.OriginalURL == req.OriginalURL
		})

		if idx == -1 {
			http.Error(rw, "internal error", http.StatusInternalServerError)
			return
		}
		reply[i].CorrelationID = request[idx].CorrelationID
		reply[i].ShortURL = link.ShortURL
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(rw)
	enc.SetEscapeHTML(false)
	err = enc.Encode(reply)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GET /api/user/urls
func (h *Handler) getUserURLs(rw http.ResponseWriter, req *http.Request) {
	uid := h.getUserID(req)
	links, err := h.shortener.GetLinksByUserID(uid)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if len(links) == 0 {
		rw.WriteHeader(http.StatusNoContent)
		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(rw)
	enc.SetEscapeHTML(false)
	err = enc.Encode(links)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DELETE /api/user/urls
func (h *Handler) deleteURLs(rw http.ResponseWriter, req *http.Request) {
	var shortURLs []string
	if err := json.NewDecoder(req.Body).Decode(&shortURLs); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	userID := h.getUserID(req)
	h.workers.Go(func() error {
		return h.shortener.DeleteShortURLs(userID, shortURLs)
	})
	rw.WriteHeader(http.StatusAccepted)
}

// GET /ping
func (h *Handler) ping(rw http.ResponseWriter, _ *http.Request) {
	if err := h.shortener.Ping(); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

// shorten Cократить ссылку.
// Ф-ция так же укстанавлиает cookie "user_id".
func (h *Handler) shorten(req *http.Request, rw http.ResponseWriter, originalURL string) (model.Link, error) {
	userID := h.getUserID(req)

	link, err := h.shortener.CreateLink(&userID, originalURL)
	if err != nil && !errors.Is(err, model.ErrLinkAlreadyExists) {
		return model.Link{}, err
	}

	h.setUserID(rw, userID)
	return link, err
}

func (h *Handler) shortenBatch(req *http.Request, rw http.ResponseWriter, originalURLs []string) ([]model.Link, error) {
	userID := h.getUserID(req)

	links, err := h.shortener.CreateLinks(&userID, originalURLs)
	if err != nil {
		return nil, err
	}

	h.setUserID(rw, userID)
	return links, nil
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
			_ = bodyDecompressor.Close()
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
