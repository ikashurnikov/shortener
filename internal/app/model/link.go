package model

import (
	"net/url"
)

type LinkID uint32

type Link struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NormalizeOriginalURL(originalURL string) (string, error) {
	if originalURL == "" {
		return "", ErrInvalidURL
	}
	u, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return "", ErrInvalidURL
	}
	return u.String(), nil
}

func NormalizeOriginalURLs(originalURLs []string) ([]string, error) {
	if len(originalURLs) == 0 {
		return nil, nil
	}

	res := make([]string, len(originalURLs))
	for i, u := range originalURLs {
		normLink, err := NormalizeOriginalURL(u)
		if err != nil {
			return nil, err
		}
		res[i] = normLink
	}

	return res, nil
}
