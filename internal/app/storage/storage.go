package storage

import "errors"

type Storage interface {
	Insert(value string) (uint32, error)
	Select(id uint32) (string, error)
	Close() error
}

var (
	ErrNotFound = errors.New("not found")
)
