package model

import (
	"errors"
	"github.com/ikashurnikov/shortener/internal/app/storage"
	"github.com/ikashurnikov/shortener/internal/app/urlencoder"
	"net/url"
	"strings"
)

type DefaultModel struct {
	urlEncoder     urlencoder.Encoder
	storage        storage.Storage
	shortURLPrefix string
}

func New(storage storage.Storage, baseURL url.URL) *DefaultModel {
	shortURLPrefix := baseURL.String()
	if !strings.HasSuffix(shortURLPrefix, "/") {
		shortURLPrefix = shortURLPrefix + "/"
	}

	return &DefaultModel{
		storage:        storage,
		urlEncoder:     urlencoder.NewZBase32Encoder(),
		shortURLPrefix: shortURLPrefix,
	}
}

func (m *DefaultModel) ShortenLink(userID *UserID, link string) (string, error) {
	normLink, err := NormalizeLink(link)
	if err != nil {
		return "", err
	}

	if err := m.addUser(userID); err != nil {
		return "", err
	}

	linkID, err := m.storage.InsertLink(storage.UserID(*userID), normLink)
	if err != nil {
		return "", err
	}

	shortLink, err := m.makeShortURL(linkID)
	if err != nil {
		return "", err
	}

	return shortLink, nil
}

func (m *DefaultModel) ShortenLinks(userID *UserID, links []string) ([]string, error) {
	normLinks, err := NormalizeLinks(links)
	if err != nil {
		return nil, err
	}

	if err := m.addUser(userID); err != nil {
		return nil, err
	}

	linkIDs, err := m.storage.InsertLinks(storage.UserID(*userID), normLinks)
	if err != nil {
		return nil, err
	}

	res := make([]string, len(linkIDs))
	for i, linkID := range linkIDs {
		shortLink, err := m.makeShortURL(linkID)
		if err != nil {
			return nil, err
		}
		res[i] = shortLink
	}
	return res, err
}

func (m *DefaultModel) ExpandLink(shortURL string) (string, error) {
	id, err := m.urlEncoder.Expand(shortURL)
	if err != nil {
		return "", err
	}
	return m.storage.SelectLink(storage.LinkID(id))
}

func (m *DefaultModel) UserLinks(id UserID) (map[string]string, error) {
	if id == InvalidUserID {
		return nil, nil
	}

	links, err := m.storage.SelectUserLinks(storage.UserID(id))
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, nil
		}
		return nil, err
	}

	res := make(map[string]string)
	for longLink, id := range links {
		shortLink, err := m.makeShortURL(id)
		if err != nil {
			return nil, err
		}
		res[longLink] = shortLink
	}
	return res, nil
}

func (m *DefaultModel) Ping() error {
	return m.storage.Ping()
}

func (m *DefaultModel) Close() error {
	return m.storage.Close()
}

func (m *DefaultModel) makeShortURL(id storage.LinkID) (string, error) {
	shortLink, err := m.urlEncoder.Shorten(uint32(id))
	if err != nil {
		return "", err
	}
	return m.shortURLPrefix + shortLink, nil
}

func (m *DefaultModel) addUser(id *UserID) error {
	if id == nil {
		return errors.New("userID is null")
	}

	if *id == InvalidUserID {
		userID, err := m.storage.InsertUser()
		if err != nil {
			return err
		}
		*id = UserID(userID)
	}

	return nil
}

func NormalizeLink(link string) (string, error) {
	if link == "" {
		return "", errors.New("link is empty")
	}
	u, err := url.ParseRequestURI(link)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func NormalizeLinks(links []string) ([]string, error) {
	if len(links) == 0 {
		return nil, nil
	}

	res := make([]string, len(links))
	for i, link := range links {
		normLink, err := NormalizeLink(link)
		if err != nil {
			return nil, err
		}
		res[i] = normLink
	}

	return res, nil
}
