package urlshortener

import (
	"github.com/ikashurnikov/shortener/internal/app/storage"
	"github.com/ikashurnikov/shortener/internal/app/str2int"
	"net/url"
)

type StdShortener struct {
	Encoder str2int.Encoder
	Storage storage.Storage
}

func (short *StdShortener) EncodeLongURL(longURL string) (string, error) {
	url, err := url.ParseRequestURI(longURL)
	if err != nil {
		return "", err
	}
	id, err := short.Storage.Insert(url.String())
	if err != nil {
		return "", err
	}
	return short.Encoder.EncodeToString(id)
}

func (short *StdShortener) DecodeShortURL(shortURL string) (string, error) {
	id, err := short.Encoder.DecodeString(shortURL)
	if err != nil {
		return "", err
	}
	return short.Storage.Select(id)
}
