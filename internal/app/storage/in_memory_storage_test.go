package storage

import (
	"testing"
)

func TestInMemoryStorage(t *testing.T) {
	testStorage(func() Storage {
		return NewInMemoryStorage()
	}, t)
}
