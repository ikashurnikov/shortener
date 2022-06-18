package storage

import (
	"sync"
)

type InMemoryStorage struct {
	id2value  map[uint32]string
	value2id  map[string]uint32
	currentID uint32
	guard     sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		id2value:  make(map[uint32]string),
		value2id:  make(map[string]uint32),
		currentID: 1,
	}
}

func (s *InMemoryStorage) Select(id uint32) (string, error) {
	s.guard.RLock()
	defer s.guard.RUnlock()

	value, ok := s.id2value[id]
	if ok {
		return value, nil
	}
	return "", ErrNotFound
}

func (s *InMemoryStorage) Insert(value string) (uint32, error) {
	s.guard.Lock()
	defer s.guard.Unlock()

	id, ok := s.value2id[value]
	if ok {
		return id, nil
	}

	id = s.currentID
	s.id2value[id] = value
	s.value2id[value] = id
	s.currentID++

	return id, nil
}

func (s *InMemoryStorage) Close() error {
	return nil
}
