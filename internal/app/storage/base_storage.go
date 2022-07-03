package storage

import (
	"github.com/ikashurnikov/shortener/internal/app/urlencoder"
	"net/url"
)

type (
	newURLInfoFunc = func(id uint32, longURL string) (URLInfo, error)

	// storageImpl Интерфейс реализации хранилища
	// Аргументы передаваемые в интерфейс уже прошли все проверки.
	storageImpl interface {
		insertLongURL(userID *UserID, url string) (uint32, error)
		selectLongURL(id uint32) (string, error)
		getUserURLs(id UserID, newURLInfo newURLInfoFunc) ([]URLInfo, error)
		ping() error
		close() error
	}

	// baseStorage Базовая реализация хранилища.
	// Выполняет необходимые проверки входящих аргументов и перенаправляет вызов в storageImpl
	baseStorage struct {
		urlEncoder urlencoder.Encoder
		storageImpl
	}
)

func (s *baseStorage) AddLongURL(userID *UserID, longURL string, baseURL url.URL) (string, error) {
	normLongURL, err := NormalizeLongURL(longURL)
	if err != nil {
		return "", err
	}

	id, err := s.insertLongURL(userID, normLongURL)
	if err != nil {
		return "", err
	}

	shortURL, err := s.urlEncoder.Shorten(id)
	if err != nil {
		return "", ErrEncodingLongURL
	}
	baseURL.Path = shortURL
	return baseURL.String(), nil
}

func (s *baseStorage) GetLongURL(shortURL string) (string, error) {
	if shortURL == "" {
		return "", ErrDecodingShortURL
	}

	longURL, err := s.urlEncoder.Expand(shortURL)
	if err != nil {
		return "", ErrDecodingShortURL
	}

	return s.selectLongURL(longURL)
}

func (s *baseStorage) GetUserURLs(userID UserID, baseURL url.URL) ([]URLInfo, error) {
	if err := IsValidUserID(userID); err != nil {
		return nil, ErrUserNotFound
	}

	return s.getUserURLs(userID,
		func(id uint32, longURL string) (URLInfo, error) {
			uinfo, err := s.newURLInfo(id, longURL)
			if err != nil {
				return uinfo, err
			}
			baseURL.Path = uinfo.ShortURL
			uinfo.ShortURL = baseURL.String()
			return uinfo, nil
		})
}

func (s *baseStorage) Ping() error {
	return s.ping()
}

func (s *baseStorage) Close() error {
	return s.close()
}

func (s *baseStorage) newURLInfo(id uint32, longURL string) (URLInfo, error) {
	shortURL, err := s.urlEncoder.Shorten(id)
	if err != nil {
		return URLInfo{}, ErrEncodingLongURL
	}
	return URLInfo{LongURL: longURL, ShortURL: shortURL}, nil
}
