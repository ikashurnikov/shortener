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

	insert := func(value string, want_id uint32) {
		id, err := storage.Insert(value)
		require.NoError(t, err)
		require.Equal(t, want_id, id)
	}

	select_ := func(id uint32, want_value string) {
		value, err := storage.Select(id)
		require.NoError(t, err)
		require.Equal(t, want_value, value)
	}

	insert("one", 1)
	select_(1, "one")
	insert("two", 2)
	select_(2, "two")
	insert("one", 1)
	storage.Close()

	storage, err = NewFileStorage(filename)
	require.NoError(t, err)
	select_(1, "one")
	select_(2, "two")
	insert("two", 2)
	insert("three", 3)
}
