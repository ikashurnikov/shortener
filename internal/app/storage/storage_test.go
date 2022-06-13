package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testInsertFunc = func(storage Storage, t *testing.T)
type testSelectFunc = func(storage Storage, t *testing.T)

var (
	testInsert testInsertFunc = func(storage Storage, t *testing.T) {
		defer storage.Close()

		tests := []struct {
			name  string
			value string
			want  uint32
		}{
			{
				name:  "1: insert",
				value: "value1",
				want:  1,
			},
			{
				name:  "2: insert",
				value: "value2",
				want:  2,
			},
			{
				name:  "1: insert duplicate",
				value: "value1",
				want:  1,
			},
			{
				name:  "2: insert duplicate",
				value: "value2",
				want:  2,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := storage.Insert(tt.value)
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			})
		}
	}

	testSelect testSelectFunc = func(storage Storage, t *testing.T) {
		defer storage.Close()

		tests := []struct {
			name    string
			id      uint32
			want    string
			wantErr bool
		}{
			{
				name:    "select existing id",
				id:      1,
				want:    "value1",
				wantErr: false,
			},
			{
				name:    "select unknown id",
				id:      2,
				want:    "",
				wantErr: true,
			},
		}

		id, err := storage.Insert("value1")
		require.NoError(t, err)
		require.Equal(t, id, uint32(1))

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := storage.Select(tt.id)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.want, got)
				}
			})
		}
	}
)
