package service

import "github.com/ikashurnikov/shortener/internal/app/model"

type Shortener interface {
	CreateLink(userID *model.UserID, originalURL string) (model.Link, error)
	CreateLinks(userID *model.UserID, originalURLs []string) ([]model.Link, error)
	GetLinkByShortURL(shortURL string) (model.Link, error)
	GetLinksByUserID(id model.UserID) ([]model.Link, error)
	Ping() error
}

type LinkIDEncoder interface {
	EncodeToString(id model.LinkID) (string, error)
	DecodeFromString(str string) (model.LinkID, error)
}
