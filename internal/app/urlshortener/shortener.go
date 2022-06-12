package urlshortener

import "net/url"

type Shortener interface {
	EncodeLongURL(longURL string) (string, error)
	DecodeShortURL(shortURL string) (string, error)
}

func ShortenURL(shortener Shortener, longURL string, baseShortURL *url.URL) error {
	path, err := shortener.EncodeLongURL(longURL)
	if err != nil {
		return err
	}

	baseShortURL.Path = path
	return nil
}
