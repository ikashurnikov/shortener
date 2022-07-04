package storage

import (
	"errors"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrLinkNotFound = errors.New("link not found")
)

const (
	InvalidUserID UserID = -1
)

type (
	UserID int32
	LinkID uint32

	Storage interface {
		//InsertUser Добавляет нового пользователя
		InsertUser() (UserID, error)

		//InsertLink  Добавляет ссылку и возвращает ее ID
		InsertLink(userID UserID, link string) (LinkID, error)

		//InsertLinks Добвляет ссылки и возвращает их ID
		InsertLinks(userID UserID, links []string) ([]LinkID, error)

		//SelectLink Возвращает ссылку по ее ID
		SelectLink(id LinkID) (string, error)

		//SelectUserLinks возвращает ссылки и их ID, привязанные к пользователю.
		//Если пользователя не существует, возвращает пустую карту
		SelectUserLinks(id UserID) (map[string]LinkID, error)

		Ping() error

		Close() error
	}
)
