package storage

import (
	"fmt"
	"sync"
)

type InMemoryStorage struct {
	id2value  map[uint32]string
	value2id  map[string]uint32
	currentID uint32
	guard     sync.RWMutex
}

func NewInMemoryStorage() Storage {
	storage := new(InMemoryStorage)
	storage.id2value = make(map[uint32]string)
	storage.value2id = make(map[string]uint32)
	storage.currentID = 1
	return storage
}

func (storage *InMemoryStorage) Select(id uint32) (string, error) {
	storage.guard.RLock()
	defer storage.guard.RUnlock()

	value, ok := storage.id2value[id]
	if ok {
		return value, nil
	}
	return "", fmt.Errorf("id '%d' not found", id)
}

func (storage *InMemoryStorage) Insert(value string) (uint32, error) {
	storage.guard.Lock()
	defer storage.guard.Unlock()

	id, ok := storage.value2id[value]
	if ok {
		return id, nil
	}

	id = storage.currentID
	storage.id2value[id] = value
	storage.value2id[value] = id
	storage.currentID++

	return id, nil
}
