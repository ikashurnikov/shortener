package service

import (
	"fmt"
	"github.com/ikashurnikov/shortener/internal/app/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZBase32LinkIDEncoder_DecodeFromString(t *testing.T) {
	tests := []struct {
		name    string
		str     string
		want    model.LinkID
		wantErr bool
	}{
		{
			name:    "1. decoding correct string",
			str:     "yyyyyyy",
			want:    0,
			wantErr: false,
		},
		{
			name:    "2. decoding correct string",
			str:     "999999a",
			want:    0xffffffff,
			wantErr: false,
		},
		{
			name:    "decoding too short string",
			str:     "nre",
			want:    0,
			wantErr: true,
		},
		{
			name:    "decoding too long string",
			str:     "nreoxsdsd2",
			want:    0,
			wantErr: true,
		},
		{
			name:    "decoding string with invalid symbols",
			str:     "[][]",
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoder := NewZBase32LinkIDEncoder()
			got, err := encoder.DecodeFromString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestZBase32LinkIDEncoder_EncodeToString(t *testing.T) {
	tests := []struct {
		value model.LinkID
		want  string
	}{
		{
			value: 0,
			want:  "yyyyyyy",
		},
		{
			value: 0xffffffff,
			want:  "999999a",
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("encodig %v", tt.value)
		t.Run(name, func(t *testing.T) {
			encoder := NewZBase32LinkIDEncoder()
			got, err := encoder.EncodeToString(tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
