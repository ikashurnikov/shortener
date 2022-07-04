package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ikashurnikov/shortener/internal/app/model"
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

type mockModel struct {
	error   bool
	baseURL url.URL
	userID  model.UserID
}

func newMockModel(err bool, baseURL url.URL) *mockModel {
	return &mockModel{
		error:   err,
		baseURL: baseURL,
	}
}

func (m *mockModel) ShortenLink(userID *model.UserID, longURL string) (string, error) {
	if m.error {
		return "", errors.New("error")
	}
	*userID = m.userID
	longURL = strings.TrimPrefix(longURL, "https://")
	shortURL := strings.TrimPrefix(longURL, "http://")
	m.baseURL.Path = shortURL
	return m.baseURL.String(), nil
}

func (m *mockModel) ShortenLinks(userID *model.UserID, urls []string) ([]string, error) {
	return nil, errors.New("not implemented")
}

func (m *mockModel) ExpandLink(shortURL string) (string, error) {
	if m.error {
		return "", errors.New("error")
	}
	m.baseURL.Path = shortURL
	return m.baseURL.String(), nil
}

func (m *mockModel) UserLinks(userID model.UserID) (map[string]string, error) {
	return nil, errors.New("not implemented")
}

func (m *mockModel) Close() error {
	return nil
}

func (m *mockModel) Ping() error {
	if m.error {
		return errors.New("error")
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

			m := newMockModel(tt.storageError, baseURL)
			handler := NewHandler(m, "secret")
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

			m := newMockModel(tt.storageError, baseURL)
			handler := NewHandler(m, "secret")
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
				location:   "http://localhost:8080/yandex.ru",
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

			m := newMockModel(tt.storageError, baseURL)
			handler := NewHandler(m, "secret")
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

			m := newMockModel(false, baseURL)
			handler := NewHandler(m, "secret")
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
