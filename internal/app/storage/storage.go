package storage

import (
	"errors"
	"net/url"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidUserID    = errors.New("invalid user id")
	ErrDecodingShortURL = errors.New("decoding short url failed")
	ErrInvalidLongURL   = errors.New("invalid long url")
	ErrEncodingLongURL  = errors.New("encoding long url failed")
	ErrStorage          = errors.New("storage error")
)

type UserID int

const (
	InvalidUserID UserID = -1
)

type URLInfo struct {
	LongURL  string `json:"original_url"`
	ShortURL string `json:"short_url"`
}

type Storage interface {
	// AddLongURL Добавить оригинальную ссылку в хранилище.
	// Возвращает сокращенную ссылку и индетификатор пользователя.
	// Если ссылка уже есть в хранилище, вернет существующую короткую ссылку.
	// Если пользователя с заданым userID не существует, создаст нового пользователя и вернет его ID.
	AddLongURL(userID *UserID, longURL string, baseURL url.URL) (string, error)

	// GetLongURL Возвращает длиную ссылку по короткой.
	GetLongURL(shortURL string) (string, error)

	// GetUserURLs Возвращает все ссылки которые принадлежат данному пользователю.
	// Если пользователья с userID не существует возвращет ErrUserNotFound
	GetUserURLs(userID UserID, baseURL url.URL) ([]URLInfo, error)

	Close() error
}

// NormalizeLongURL Проверяет валидность ссылки и возвращает ее нормализованое представление в виде строки.
func NormalizeLongURL(longURL string) (string, error) {
	if longURL == "" {
		return "", ErrInvalidLongURL
	}
	u, err := url.ParseRequestURI(longURL)
	if err != nil {
		return "", ErrInvalidLongURL
	}
	return u.String(), nil
}

func IsValidUserID(id UserID) error {
	if id < 0 {
		return ErrInvalidUserID
	}
	return nil
}
