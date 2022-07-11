package repo

import "github.com/ikashurnikov/shortener/internal/app/model"

type Repo interface {
	// AddUser Добавляет нового пользователя.
	AddUser() (model.UserID, error)

	//SaveOriginalURL  Сохраняет ссылку и возвращает ее ID.
	//Если сыылка уже была добавлена, возвращает так же ошибка ErrLinkAlreadyExists.
	SaveOriginalURL(userID model.UserID, originalURL string) (model.LinkID, error)

	// SaveOriginalURLs Сохраняет ссылки и возвращает их ID
	SaveOriginalURLs(userID model.UserID, originalURLs []string) ([]model.LinkID, error)

	// GetOriginalURLByID Возвращает ссылку по ее ID
	GetOriginalURLByID(id model.LinkID) (string, error)

	// GetOriginalURLsByUserID возвращает ссылки и их ID, привязанные к пользователю.
	// Если пользователя не существует, возвращает пустую карту
	GetOriginalURLsByUserID(id model.UserID) (map[string]model.LinkID, error)

	Ping() error

	Close() error
}
