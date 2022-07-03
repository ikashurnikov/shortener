package storage

import (
	"context"
	"database/sql"
	"github.com/ikashurnikov/shortener/internal/app/urlencoder"
	"time"

	_ "github.com/lib/pq"
)

type DBStorage struct {
	db *sql.DB

	baseStorage
}

func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	s := new(DBStorage)
	s.db = db
	s.urlEncoder = urlencoder.NewZBase32Encoder()
	s.baseStorage.storageImpl = s

	return s, nil
}

func (s *DBStorage) insertLongURL(userID *UserID, url string) (uint32, error) {
	return 0, ErrStorage
}

func (s *DBStorage) selectLongURL(id uint32) (string, error) {
	return "", ErrStorage
}

func (s *DBStorage) getUserURLs(id UserID, newURLInfo newURLInfoFunc) ([]URLInfo, error) {
	return nil, ErrStorage
}

func (s *DBStorage) ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return ErrStorage
	}
	return nil
}

func (s *DBStorage) close() error {
	return s.db.Close()
}
