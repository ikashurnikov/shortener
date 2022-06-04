package str2int

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZBase32_DecodeString(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		str     string
		want    uint32
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
			encoder := NewZBase32Encoder()
			got, err := encoder.DecodeString(tt.str)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestZBase32_EncodeToString(t *testing.T) {
	tests := []struct {
		value uint32
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
			encoder := NewZBase32Encoder()
			got, err := encoder.EncodeToString(tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
