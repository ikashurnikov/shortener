package storage

import (
	"io"
	"os"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

type FileStorage struct {
	file      *os.File
	encoder   *msgpack.Encoder
	decoder   *msgpack.Decoder
	currentID uint32
	guard     sync.Mutex
}

type record struct {
	id    uint32
	value string
}

func NewFileStorage(filename string) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return nil, err
	}
	storage := &FileStorage{
		file:      file,
		encoder:   msgpack.NewEncoder(file),
		decoder:   msgpack.NewDecoder(file),
		currentID: 0,
	}

	err = storage.scan(func(rec record) bool {
		storage.currentID = rec.id
		return false
	})
	if err != io.EOF {
		return nil, err
	}
	return storage, nil
}

func (s *FileStorage) scan(handler func(rec record) bool) error {
	_, err := s.file.Seek(0, 0)
	if err != nil {
		return err
	}
	s.decoder.Reset(s.file)

	var id uint32 = 1
	for {
		value, err := s.decoder.DecodeString()
		if err != nil {
			return err
		}
		if handler(record{id: id, value: value}) {
			return nil
		}
		id++
	}
}

func (s *FileStorage) Select(id uint32) (string, error) {
	s.guard.Lock()
	defer s.guard.Unlock()

	var value string
	err := s.scan(func(rec record) bool {
		if id == rec.id {
			value = rec.value
			return true
		}
		return false
	})

	if err == io.EOF {
		err = ErrNotFound
	}

	return value, err
}

func (s *FileStorage) Insert(value string) (uint32, error) {
	s.guard.Lock()
	defer s.guard.Unlock()

	var id uint32
	err := s.scan(func(rec record) bool {
		if value == rec.value {
			id = rec.id
			return true
		}
		return false
	})

	if err == nil {
		return id, nil
	} else if err != io.EOF {
		return 0, ErrNotFound
	}

	_, err = s.file.Seek(0, 2)
	if err != nil {
		return 0, err
	}

	err = s.encoder.EncodeString(value)
	if err != nil {
		return 0, err
	}

	s.currentID++
	return s.currentID, nil
}

func (s *FileStorage) Close() error {
	return s.file.Close()
}
