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

func (db *InMemoryStorage) Select(id uint32) (string, error) {
	db.guard.RLock()
	defer db.guard.RUnlock()

	value, ok := db.id2value[id]
	if ok {
		return value, nil
	}
	return "", fmt.Errorf("id '%d' not found", id)
}

func (db *InMemoryStorage) Insert(value string) (uint32, error) {
	db.guard.Lock()
	defer db.guard.Unlock()

	id, ok := db.value2id[value]
	if ok {
		return id, nil
	}

	id = db.currentID
	db.id2value[id] = value
	db.value2id[value] = id
	db.currentID++

	return id, nil
}
