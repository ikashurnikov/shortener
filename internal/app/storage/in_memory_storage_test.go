package storage

import (
	"testing"
)

func TestInMemoryStorage_Insert(t *testing.T) {
	testInsert(NewInMemoryStorage(), t)
}

func TestInMemoryStorage_Select(t *testing.T) {
	testSelect(NewInMemoryStorage(), t)
}
