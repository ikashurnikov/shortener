package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ikashurnikov/shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type mockStorage struct {
	error  bool
	userID storage.UserID
	urls   []storage.URLInfo
}

func (s *mockStorage) AddLongURL(userID *storage.UserID, longURL string, baseURL url.URL) (string, error) {
	if s.error {
		return "", storage.ErrStorage
	}
	*userID = s.userID
	longURL = strings.TrimPrefix(longURL, "https://")
	shortURL := strings.TrimPrefix(longURL, "http://")
	baseURL.Path = shortURL
	return baseURL.String(), nil
}

func (s *mockStorage) GetLongURL(shortURL string) (string, error) {
	if s.error {
		return "", storage.ErrStorage
	}
	return fmt.Sprintf("http://%v", shortURL), nil
}

func (s *mockStorage) GetUserURLs(userID storage.UserID, baseURL url.URL) ([]storage.URLInfo, error) {
	if s.error {
		return nil, storage.ErrStorage
	}
	if s.userID != userID {
		return nil, storage.ErrUserNotFound
	}
	return s.urls, nil
}

func (s *mockStorage) Close() error {
	return nil
}

func (s *mockStorage) Ping() error {
	if s.error {
		return storage.ErrStorage
	}
	return nil
}

const (
	cipherKey = "test_secret"
)

type request struct {
	method string
	path   string
	body   string

	userID storage.UserID
}

func testRequest(t *testing.T, ts *httptest.Server, r request) (*http.Response, string) {
	req, err := http.NewRequest(r.method, ts.URL+r.path, bytes.NewReader([]byte(r.body)))
	require.NoError(t, err)

	if r.userID != storage.InvalidUserID {
		cookie := NewSignedCookie(cipherKey)
		cookie.AddInt(req, "user_uid", int(r.userID))
	}

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
		name         string
		target       string
		body         string
		storageError bool
		want         want
	}{
		{
			name:         "create short link",
			target:       "/",
			body:         "https://yandex.ru",
			storageError: false,
			want: want{
				statusCode: http.StatusCreated,
				response:   "http://localhost:8080/yandex.ru",
			},
		},

		{
			name:         "invalid target",
			target:       "/xxx",
			body:         "https://yandex.ru",
			storageError: false,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},

		{
			name:         "url_shortener failed",
			target:       "/",
			body:         "https://yandex.ru",
			storageError: true,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL := url.URL{
				Scheme: "http",
				Host:   "localhost:8080",
			}

			handler := NewHandler(&mockStorage{error: tt.storageError}, baseURL, "secret")
			testServer := httptest.NewServer(handler)
			defer testServer.Close()

			resp, respBody := testRequest(t, testServer, request{
				method: "POST",
				path:   tt.target,
				body:   tt.body,
				userID: storage.InvalidUserID,
			})
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			if resp.StatusCode == http.StatusCreated {
				assert.Equal(t, tt.want.response, string(respBody))
			}
		})
	}
}

func Test_postAPIShorten(t *testing.T) {
	type want struct {
		response   string
		statusCode int
	}
	tests := []struct {
		name         string
		target       string
		body         string
		storageError bool
		want         want
	}{
		{
			name:         "create short link",
			target:       "/api/shorten",
			body:         `{"url": "https://yandex.ru"}`,
			storageError: false,
			want: want{
				statusCode: http.StatusCreated,
				response:   "http://localhost:8080/yandex.ru",
			},
		},

		{
			name:         "empty url",
			target:       "/api/shorten",
			body:         `{"url": ""}`,
			storageError: false,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},

		{
			name:         "bad json body",
			target:       "/api/shorten",
			body:         `{"url": 1}`,
			storageError: false,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},

		{
			name:         "invalid json",
			target:       "/api/shorten",
			body:         `{"url": 1`,
			storageError: false,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},

		{
			name:         "url_shortener failed",
			target:       "/api/shorten",
			body:         `{"url": "https://yandex.ru"}`,
			storageError: true,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL := url.URL{
				Scheme: "http",
				Host:   "localhost:8080",
			}

			handler := NewHandler(&mockStorage{error: tt.storageError}, baseURL, "secret")
			testServer := httptest.NewServer(handler)
			defer testServer.Close()

			resp, respBody := testRequest(t, testServer, request{
				method: "POST",
				path:   tt.target,
				body:   tt.body,
				userID: storage.InvalidUserID,
			})
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			if resp.StatusCode == http.StatusCreated {
				type Response struct {
					Result string `json:"result"`
				}

				var response Response
				err := json.Unmarshal([]byte(respBody), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, response.Result)
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
		name         string
		path         string
		storageError bool
		want         want
	}{
		{
			name:         "empty path",
			path:         "/",
			storageError: false,
			want: want{
				statusCode: http.StatusMethodNotAllowed,
				location:   "",
			},
		},

		{
			name:         "unknown short url",
			path:         "/xxxxxssddf",
			storageError: true,
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},

		{
			name:         "redirect",
			path:         "/yandex.ru",
			storageError: false,
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
			baseURL := url.URL{
				Scheme: "http",
				Host:   "localhost:8080",
			}

			handler := NewHandler(&mockStorage{error: tt.storageError}, baseURL, "secret")
			testServer := httptest.NewServer(handler)
			defer testServer.Close()

			resp, _ := testRequest(t, testServer, request{
				method: "GET",
				path:   tt.path,
				userID: storage.InvalidUserID,
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
			baseURL := url.URL{
				Scheme: "http",
				Host:   "localhost:8080",
			}

			handler := NewHandler(&mockStorage{}, baseURL, "secret")
			testServer := httptest.NewServer(handler)
			defer testServer.Close()

			resp, _ := testRequest(t, testServer, request{
				method: tt.method,
				path:   tt.path,
				body:   tt.body,
				userID: storage.InvalidUserID,
			})
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
		})
	}
}
