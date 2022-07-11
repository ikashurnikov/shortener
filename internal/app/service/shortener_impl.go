package service

import (
	"errors"
	"github.com/ikashurnikov/shortener/internal/app/model"
	"github.com/ikashurnikov/shortener/internal/app/repo"
	"net/url"
	"strings"
)

type shortener struct {
	linkIDEncoder  LinkIDEncoder
	repo           repo.Repo
	shortURLPrefix string
}

func NewShortener(repo repo.Repo, baseURL url.URL) *shortener {
	shortURLPrefix := baseURL.String()
	if !strings.HasSuffix(shortURLPrefix, "/") {
		shortURLPrefix = shortURLPrefix + "/"
	}

	return &shortener{
		repo:           repo,
		linkIDEncoder:  NewZBase32LinkIDEncoder(),
		shortURLPrefix: shortURLPrefix,
	}
}

func (s *shortener) CreateLink(userID *model.UserID, originalURL string) (model.Link, error) {
	originalURL, err := model.NormalizeOriginalURL(originalURL)
	if err != nil {
		return model.Link{}, err
	}

	if err = s.addUser(userID); err != nil {
		return model.Link{}, err
	}

	linkID, err := s.repo.SaveOriginalURL(*userID, originalURL)
	if err != nil && !errors.Is(err, model.ErrLinkAlreadyExists) {
		return model.Link{}, err
	}

	link, err2 := s.createLink(linkID, originalURL)
	if err2 != nil {
		return model.Link{}, err2
	}
	return link, err
}

func (s *shortener) CreateLinks(userID *model.UserID, originalURLs []string) ([]model.Link, error) {
	originalURLs, err := model.NormalizeOriginalURLs(originalURLs)
	if err != nil {
		return nil, err
	}

	if err := s.addUser(userID); err != nil {
		return nil, err
	}

	linkIDs, err := s.repo.SaveOriginalURLs(*userID, originalURLs)
	if err != nil {
		return nil, err
	}

	res := make([]model.Link, 0, len(linkIDs))
	for idx, linkID := range linkIDs {
		link, err := s.createLink(linkID, originalURLs[idx])
		if err != nil {
			return nil, err
		}
		res = append(res, link)
	}

	return res, nil
}

func (s *shortener) GetLinkByShortURL(shortURL string) (model.Link, error) {
	linkID, err := s.linkIDEncoder.DecodeFromString(shortURL)
	if err != nil {
		return model.Link{}, err
	}

	origURL, err := s.repo.GetOriginalURLByID(linkID)
	if err != nil {
		return model.Link{}, err
	}

	return s.createLink(linkID, origURL)
}

func (s *shortener) GetLinksByUserID(userID model.UserID) ([]model.Link, error) {
	if !userID.IsValid() {
		return nil, nil
	}

	origURLToLinkIDs, err := s.repo.GetOriginalURLsByUserID(userID)
	if err != nil {
		if errors.Is(err, model.ErrUserNotFound) {
			return nil, nil
		}
		return nil, err
	}

	res := make([]model.Link, 0, len(origURLToLinkIDs))
	for origURL, linkID := range origURLToLinkIDs {
		link, err := s.createLink(linkID, origURL)
		if err != nil {
			return nil, err
		}
		res = append(res, link)
	}

	return res, err
}

func (s *shortener) Ping() error {
	return s.repo.Ping()
}

func (s *shortener) addUser(userID *model.UserID) error {
	if userID == nil {
		return model.ErrInvalidUserID
	}

	if !userID.IsValid() {
		var err error
		*userID, err = s.repo.AddUser()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *shortener) createLink(linkID model.LinkID, originalURL string) (model.Link, error) {
	shortURL, err := s.createShortURL(linkID)
	return model.Link{OriginalURL: originalURL, ShortURL: shortURL}, err
}

func (s *shortener) createShortURL(id model.LinkID) (string, error) {
	shortURL, err := s.linkIDEncoder.EncodeToString(id)
	if err != nil {
		return "", err
	}
	return s.shortURLPrefix + shortURL, nil
}
