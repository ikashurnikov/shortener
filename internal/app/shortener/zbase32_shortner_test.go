package shortener

import (
	"github.com/ikashurnikov/shortener/internal/app/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestZBase32Shortener_Decode(t *testing.T) {
	shortener := NewZBase32Shortener(storage.NewInMemoryStorage())

	longURLS := [2]string{
		"http://yandex.ru",
		"http://google.com",
	}
	shortURLs := [2]string{}

	for i, longURL := range longURLS {
		shortURL, err := shortener.Encode(longURL)
		require.NoError(t, err)
		require.NotEqual(t, "", shortURL)
		shortURLs[i] = shortURL
	}

	tests := []struct {
		name     string
		shortURL string
		want     string
		wantErr  bool
	}{
		{
			name:     "1: decode existing url ",
			shortURL: shortURLs[0],
			want:     longURLS[0],
			wantErr:  false,
		},
		{
			name:     "2: decode existing url",
			shortURL: shortURLs[1],
			want:     longURLS[1],
			wantErr:  false,
		},
		{
			name:     "decoding unknown url",
			shortURL: "1234567",
			wantErr:  true,
		},
		{
			name:     "decoding unknown long url",
			shortURL: "1234567wewewewewe",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shortener.Decode(tt.shortURL)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestZBase32Shortener_Encode(t *testing.T) {
	shortener := NewZBase32Shortener(storage.NewInMemoryStorage())
	yandexShortURL, err := shortener.Encode("http://yandex.ru")
	assert.NoError(t, err)
	assert.NotEqual(t, "", yandexShortURL)

	googleShortURL, err := shortener.Encode("http://google.com")
	assert.NoError(t, err)
	assert.NotEqual(t, "", googleShortURL)

	assert.NotEqual(t, yandexShortURL, googleShortURL)

	yandexShortURL2, err := shortener.Encode("http://yandex.ru")
	assert.NoError(t, err)
	assert.Equal(t, yandexShortURL, yandexShortURL2)

	googleShortURL2, err := shortener.Encode("http://google.com")
	assert.NoError(t, err)
	assert.Equal(t, googleShortURL, googleShortURL2)
}
