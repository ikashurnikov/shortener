package handler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type mockShortener struct {
	error bool
}

func (shortener *mockShortener) EncodeLongURL(longURL string) (string, error) {
	if shortener.error {
		return "", errors.New("ERROR")
	}
	return strings.TrimPrefix(longURL, "http://"), nil
}

func (shortener *mockShortener) DecodeShortURL(shortURL string) (string, error) {
	if shortener.error {
		return "", errors.New("ERROR")
	}
	return fmt.Sprintf("http://%v", shortURL), nil
}

type request struct {
	method string
	path   string
	body   string
}

func testRequest(t *testing.T, ts *httptest.Server, r request) (*http.Response, string) {
	req, err := http.NewRequest(r.method, ts.URL+r.path, bytes.NewReader([]byte(r.body)))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func Test_postLongLink(t *testing.T) {
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
				response:   "http://localhost:8080/yandex.ru",
			},
		},
		{
			name:           "invalid target",
			target:         "/xxx",
			body:           "http://yandex.ru",
			shortenerError: false,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},

		{
			name:           "url_shortener failed",
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
			baseURL := url.URL{
				Scheme: "http",
				Host:   "localhost:8080",
			}

			handler := NewHandler(shortener, baseURL)
			testServer := httptest.NewServer(handler)
			defer testServer.Close()

			resp, respBody := testRequest(t, testServer, request{
				method: "POST",
				path:   tt.target,
				body:   tt.body,
			})
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			if resp.StatusCode == http.StatusCreated {
				assert.Equal(t, tt.want.response, string(respBody))
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
				statusCode: http.StatusMethodNotAllowed,
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

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shortener := &mockShortener{error: tt.shortenerError}
			baseURL := url.URL{
				Scheme: "http",
				Host:   "localhost:8080",
			}

			handler := NewHandler(shortener, baseURL)
			testServer := httptest.NewServer(handler)
			defer testServer.Close()

			resp, _ := testRequest(t, testServer, request{
				method: "GET",
				path:   tt.path,
			})
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			assert.Equal(t, tt.want.location, resp.Header.Get("Location"))
		})
	}
}

func Test_route(t *testing.T) {
	tests := []struct {
		method         string
		path           string
		body           string
		wantStatusCode int
	}{
		{
			method:         "POST",
			path:           "/",
			body:           "http://yandex.ru",
			wantStatusCode: http.StatusCreated,
		},
		{
			method:         "POST",
			path:           "/test",
			body:           "http://yandex.ru",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			method:         "GET",
			path:           "/xxx",
			wantStatusCode: http.StatusTemporaryRedirect,
		},
		{
			method:         "GET",
			path:           "/",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			method:         "PUT",
			path:           "/",
			body:           "http://yandex.ru",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			method:         "PATCH",
			path:           "/",
			body:           "http://yandex.ru",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			method:         "DELETE",
			body:           "http://yandex.ru",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			method:         "CONNECT",
			path:           "/",
			body:           "http://yandex.ru",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			method:         "OPTIONS",
			path:           "/",
			body:           "http://yandex.ru",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
		{
			method:         "TRACE",
			path:           "/",
			body:           "http://yandex.ru",
			wantStatusCode: http.StatusMethodNotAllowed,
		},
	}

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s %s", tt.method, tt.path)
		t.Run(name, func(t *testing.T) {
			shortener := &mockShortener{}
			baseURL := url.URL{
				Scheme: "http",
				Host:   "localhost:8080",
			}

			handler := NewHandler(shortener, baseURL)
			testServer := httptest.NewServer(handler)
			defer testServer.Close()

			resp, _ := testRequest(t, testServer, request{
				method: tt.method,
				path:   tt.path,
				body:   tt.body,
			})
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}