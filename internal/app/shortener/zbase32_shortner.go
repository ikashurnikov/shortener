package shortener

import (
	"encoding/binary"
	"fmt"
	"github.com/corvus-ch/zbase32"
	"github.com/ikashurnikov/shortener/internal/app/storage"
	"net/url"
)

type ZBase32Shortener struct {
	db storage.Storage
}

func NewZBase32Shortener() ZBase32Shortener {
	return ZBase32Shortener{
		db: storage.NewInMemoryStorage(),
	}
}

func (shortener *ZBase32Shortener) Encode(longURL string) (string, error) {
	url, err := url.Parse(longURL)
	if err != nil {
		return "", err
	}
	longURL = url.String()

	id, err := shortener.db.Insert(longURL)
	if err != nil {
		return "", err
	}

	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, id)
	return zbase32.StdEncoding.EncodeToString(bytes), nil
}

func (shortener *ZBase32Shortener) Decode(shortURL string) (string, error) {
	bytes, err := zbase32.StdEncoding.DecodeString(shortURL)
	if err != nil {
		return "", err
	}

	if len(bytes) != 4 {
		return "", fmt.Errorf("invalid short url %s", shortURL)
	}

	id := binary.LittleEndian.Uint32(bytes)
	return shortener.db.Select(id)
}
