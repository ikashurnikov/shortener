package urlshortener

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockStorage struct {
	err error
	id  uint32
	str string
}

func (storage mockStorage) Select(id uint32) (string, error) {
	return storage.str, storage.err
}

func (storage mockStorage) Insert(value string) (uint32, error) {
	return storage.id, storage.err
}

type mockEncoder struct {
	err error
	id  uint32
	str string
}

func (encoder mockEncoder) EncodeToString(value uint32) (string, error) {
	return encoder.str, encoder.err
}

func (encoder mockEncoder) DecodeString(str string) (uint32, error) {
	return encoder.id, encoder.err
}

func TestStdShortener_DecodeURL(t *testing.T) {
	tests := []struct {
		name    string
		encoder mockEncoder
		storage mockStorage
		want    string
		wantErr bool
	}{
		{
			name: "correct decoding without errors",
			encoder: mockEncoder{
				err: nil,
				id:  1,
				str: "short",
			},
			storage: mockStorage{
				err: nil,
				id:  1,
				str: "https://example.com",
			},
			want: "https://example.com",
		},
		{
			name: "storage error",
			encoder: mockEncoder{
				err: nil,
			},
			storage: mockStorage{
				err: errors.New("test"),
				str: "https://example.com",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "str2int error",
			encoder: mockEncoder{
				err: errors.New("test"),
			},
			storage: mockStorage{
				err: nil,
				str: "https://example.com",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			short := &StdShortener{
				Encoder: tt.encoder,
				Storage: tt.storage,
			}
			got, err := short.DecodeShortURL("go.by")
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStdShortener_EncodeURL(t *testing.T) {
	tests := []struct {
		name    string
		encoder mockEncoder
		storage mockStorage
		longURL string
		want    string
		wantErr bool
	}{
		{
			name:    "correct decoding without errors",
			longURL: "https://example.com",
			encoder: mockEncoder{
				err: nil,
				id:  1,
				str: "short",
			},
			storage: mockStorage{
				err: nil,
				id:  1,
				str: "https://example.com",
			},
			want:    "short",
			wantErr: false,
		},
		{
			name:    "invalid URL",
			longURL: "./example.com",
			encoder: mockEncoder{
				err: nil,
				id:  1,
				str: "short",
			},
			storage: mockStorage{
				err: nil,
				id:  1,
				str: "https://example.com",
			},
			wantErr: true,
		},
		{
			name: "storage error",
			encoder: mockEncoder{
				err: nil,
			},
			storage: mockStorage{
				err: errors.New("test"),
				str: "https://example.com",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "str2int error",
			encoder: mockEncoder{
				err: errors.New("test"),
			},
			storage: mockStorage{
				err: nil,
				str: "https://example.com",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			short := &StdShortener{
				Encoder: tt.encoder,
				Storage: tt.storage,
			}
			got, err := short.EncodeLongURL(tt.longURL)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
