package server

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type mockShortener struct {
	error bool
}

func (shortener *mockShortener) Encode(longURL string) (string, error) {
	if shortener.error {
		return "", errors.New("ERROR")
	}
	return strings.TrimPrefix(longURL, "http://"), nil
}

func (shortener *mockShortener) Decode(shortURL string) (string, error) {
	if shortener.error {
		return "", errors.New("ERROR")
	}
	return fmt.Sprintf("http://%v", shortURL), nil
}

func Test_addLongLink(t *testing.T) {
	type want struct {
		response   string
		statusCode int
	}
	tests := []struct {
		name           string
		target         string
		body           string
		shortenerError bool
		want           want
	}{
		{
			name:           "create short link",
			target:         "/",
			body:           "http://yandex.ru",
			shortenerError: false,
			want: want{
				statusCode: http.StatusCreated,
				response:   "http://localhost/yandex.ru",
			},
		},
		{
			name:           "invalid target",
			target:         "/xxx",
			body:           "http://yandex.ru",
			shortenerError: false,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},

		{
			name:           "shortener failed",
			target:         "/",
			body:           "http://yandex.ru",
			shortenerError: true,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortener := &mockShortener{error: tt.shortenerError}
			request := httptest.NewRequest(http.MethodPost, tt.target, bytes.NewReader([]byte(tt.body)))

			w := httptest.NewRecorder()
			h := http.HandlerFunc(addLongLink(shortener, url.URL{
				Scheme: "http",
				Host:   "localhost",
			}))
			h.ServeHTTP(w, request)

			res := w.Result()

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if res.StatusCode == http.StatusCreated {
				assert.Equal(t, tt.want.response, string(resBody))
			}
		})
	}
}

func Test_getShortLink(t *testing.T) {
	type want struct {
		location   string
		statusCode int
	}

	tests := []struct {
		name           string
		path           string
		shortenerError bool
		want           want
	}{
		{
			name:           "empty path",
			path:           "/",
			shortenerError: false,
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},

		{
			name:           "unknown short url",
			path:           "/xxxxxssddf",
			shortenerError: true,
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},

		{
			name:           "redirect",
			path:           "/yandex.ru",
			shortenerError: false,
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   "http://yandex.ru",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortener := &mockShortener{error: tt.shortenerError}
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(getShortLink(shortener))
			h.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.location, res.Header.Get("Location"))
		})
	}
}

func Test_route(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		wantStatusCode int
	}{
		{
			name:           "POST",
			method:         http.MethodPost,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "GET",
			method:         http.MethodGet,
			wantStatusCode: http.StatusOK,
		},

		{
			name:           "DELETE",
			method:         http.MethodDelete,
			wantStatusCode: http.StatusBadRequest,
		},

		{
			name:           "HEAD",
			method:         http.MethodHead,
			wantStatusCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "/", nil)

			handler := func(responseWriter http.ResponseWriter, request *http.Request) {
				responseWriter.WriteHeader(http.StatusOK)
			}

			w := httptest.NewRecorder()
			h := http.HandlerFunc(route(routingTable{
				"GET":  handler,
				"POST": handler,
			}))
			h.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
		})
	}
}
