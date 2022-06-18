package storage

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestFileStorage_Insert(t *testing.T) {
	filename := uuid.New().String()
	storage, err := NewFileStorage(filename)
	require.NoError(t, err)
	defer os.Remove(filename)

	testInsert(storage, t)
}

func TestFileStorage_Select(t *testing.T) {
	filename := uuid.New().String()
	storage, err := NewFileStorage(filename)
	require.NoError(t, err)
	defer os.Remove(filename)

	testSelect(storage, t)
}

func TestFileStorage_OpenClose(t *testing.T) {
	filename := uuid.New().String()
	storage, err := NewFileStorage(filename)
	require.NoError(t, err)
	defer os.Remove(filename)
	defer storage.Close()

	insertValue := func(value string, wantId uint32) {
		id, err := storage.Insert(value)
		require.NoError(t, err)
		require.Equal(t, wantId, id)
	}

	selectID := func(id uint32, wantValue string) {
		value, err := storage.Select(id)
		require.NoError(t, err)
		require.Equal(t, wantValue, value)
	}

	insertValue("one", 1)
	selectID(1, "one")
	insertValue("two", 2)
	selectID(2, "two")
	insertValue("one", 1)
	storage.Close()

	storage, err = NewFileStorage(filename)
	require.NoError(t, err)
	selectID(1, "one")
	selectID(2, "two")
	insertValue("two", 2)
	insertValue("three", 3)
}
