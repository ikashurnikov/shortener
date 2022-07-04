package storage

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
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

func (s *FileStorage) InsertUser() (UserID, error) {
	userID, err := s.cache.InsertUser()
	return userID, s.autoSave(err)
}

func (s *FileStorage) InsertLink(userID UserID, link string) (LinkID, error) {
	id, err := s.cache.InsertLink(userID, link)
	return id, s.autoSave(err)
}

func (s *FileStorage) InsertLinks(userID UserID, links []string) ([]LinkID, error) {
	ids, err := s.cache.InsertLinks(userID, links)
	return ids, s.autoSave(err)
}

func (s *FileStorage) SelectLink(id LinkID) (string, error) {
	return s.cache.SelectLink(id)
}

func (s *FileStorage) SelectUserLinks(id UserID) (map[string]LinkID, error) {
	return s.cache.SelectUserLinks(id)
}

func (s *FileStorage) autoSave(err error) error {
	if err != nil {
		return err
	}
	if s.AutoSave {
		return s.Flush()
	}
	return nil
}

func (s *FileStorage) load() error {
	stat, err := os.Stat(s.filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if stat.Size() == 0 {
		return nil
	}

	file, err := os.OpenFile(s.filename, os.O_RDONLY, 0664)
	if err != nil {
		return err
	}

	dec := msgpack.NewDecoder(file)
	if err = dec.Decode(s.cache); err != nil {
		return err
	}

	return nil
}

func (s *FileStorage) Flush() error {
	s.guard.Lock()
	defer s.guard.Unlock()

	file, err := os.OpenFile(s.filename, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}

	enc := msgpack.NewEncoder(file)
	if err = enc.Encode(s.cache); err != nil {
		file.Close()
		return err
	}

	if err = file.Close(); err != nil {
		return err
	}
	return nil
}

func (s *FileStorage) Ping() error {
	return nil
}

func (s *FileStorage) Close() error {
	if !s.AutoSave {
		return s.Flush()
	}
	return nil
}
