package model

import "errors"

const (
	InvalidUserID UserID = -1
)

type (
	UserID int

	Model interface {
		ShortenLink(userID *UserID, url string) (string, error)
		ShortenLinks(userID *UserID, urls []string) ([]string, error)
		ExpandLink(shortURL string) (string, error)
		UserLinks(id UserID) (map[string]string, error)
		Ping() error
		Close() error
	}
)

var (
	ErrLinkConflict = errors.New("link already exists")
)
