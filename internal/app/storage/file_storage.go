package storage

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"net/url"
	"os"
	"sync"
)

type FileStorage struct {
	filename string
	cache    *InMemoryStorage
	AutoSave bool
	guard    sync.Mutex
}

func NewFileStorage(filename string) (*FileStorage, error) {
	storage := &FileStorage{
		filename: filename,
		cache:    NewInMemoryStorage(),
		AutoSave: true,
	}

	if err := storage.load(); err != nil {
		return nil, err
	}
	return storage, nil
}

func (s *FileStorage) AddLongURL(userID *UserID, longURL string, baseURL url.URL) (string, error) {
	shortURL, err := s.cache.AddLongURL(userID, longURL, baseURL)
	if err == nil && s.AutoSave {
		if err = s.Flush(); err != nil {
			return "", err
		}
	}
	return shortURL, err
}

func (s *FileStorage) GetLongURL(shortURL string) (string, error) {
	return s.cache.GetLongURL(shortURL)
}

func (s *FileStorage) GetUserURLs(userID UserID, baseURL url.URL) ([]URLInfo, error) {
	return s.cache.GetUserURLs(userID, baseURL)
}

func (s *FileStorage) load() error {
	stat, err := os.Stat(s.filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return ErrStorage
	}
	if stat.Size() == 0 {
		return nil
	}

	file, err := os.OpenFile(s.filename, os.O_RDONLY, 0664)
	if err != nil {
		return ErrStorage
	}

	dec := msgpack.NewDecoder(file)
	if err = dec.Decode(s.cache); err != nil {
		return ErrStorage
	}

	return nil
}

func (s *FileStorage) Flush() error {
	s.guard.Lock()
	defer s.guard.Unlock()

	file, err := os.OpenFile(s.filename, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return ErrStorage
	}

	enc := msgpack.NewEncoder(file)
	if err = enc.Encode(s.cache); err != nil {
		file.Close()
		return ErrStorage
	}

	if err = file.Close(); err != nil {
		return ErrStorage
	}
	return nil
}

func (s *FileStorage) Close() error {
	if !s.AutoSave {
		return s.Flush()
	}
	return nil
}
