package repo

import (
	"encoding/json"
	"errors"
	"golang.org/x/exp/slices"
	"io"
	"sync"

	"github.com/ikashurnikov/shortener/internal/app/model"
)

type (
	linkDeleted bool

	item struct {
		OriginalURL string                       `json:"original_uRL"`
		Users       map[model.UserID]linkDeleted `json:"users"`
	}

	inMemoryRepo struct {
		Items      []*item      `json:"items"`
		NextUserID model.UserID `json:"next_user_id"`
		guard      sync.RWMutex
	}
)

func NewInMemoryRepo() *inMemoryRepo {
	return &inMemoryRepo{}
}

func (repo *inMemoryRepo) Serialize(w io.Writer) error {
	repo.guard.RLock()
	defer repo.guard.RUnlock()

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(repo)
}

func (repo *inMemoryRepo) Deserialize(r io.Reader) error {
	dec := json.NewDecoder(r)
	return dec.Decode(repo)
}

func (repo *inMemoryRepo) AddUser() (model.UserID, error) {
	repo.guard.Lock()
	defer repo.guard.Unlock()

	id := repo.NextUserID
	repo.NextUserID++
	return id, nil
}

func (repo *inMemoryRepo) SaveOriginalURL(userID model.UserID, originalURL string) (model.LinkID, error) {
	repo.guard.Lock()
	defer repo.guard.Unlock()

	return repo.saveOriginalURL(userID, originalURL)
}

func (repo *inMemoryRepo) SaveOriginalURLs(userID model.UserID, originalURLs []string) ([]model.LinkID, error) {
	repo.guard.Lock()
	defer repo.guard.Unlock()

	res := make([]model.LinkID, 0)
	for _, url := range originalURLs {
		id, err := repo.saveOriginalURL(userID, url)
		if err != nil && !errors.Is(err, model.ErrLinkAlreadyExists) {
			return nil, err
		}
		res = append(res, id)
	}
	return res, nil
}

func (repo *inMemoryRepo) GetOriginalURLByID(id model.LinkID) (string, error) {
	repo.guard.RLock()
	defer repo.guard.RUnlock()

	if int(id) >= len(repo.Items) {
		return "", model.ErrLinkNotFound
	}

	it := repo.Items[id]
	for _, deleted := range it.Users {
		if !deleted {
			return it.OriginalURL, nil
		}
	}

	return it.OriginalURL, model.ErrLinkRemoved
}

func (repo *inMemoryRepo) GetOriginalURLsByUserID(userID model.UserID) (map[string]model.LinkID, error) {
	repo.guard.RLock()
	defer repo.guard.RUnlock()

	if !repo.IsValidUserID(userID) {
		return nil, model.ErrUserNotFound
	}

	res := make(map[string]model.LinkID)
	for idx, it := range repo.Items {
		deleted, ok := it.Users[userID]
		if ok && !bool(deleted) {
			res[it.OriginalURL] = model.LinkID(idx)
		}
	}
	return res, nil
}

func (repo *inMemoryRepo) DeleteURLs(userID model.UserID, links []model.LinkID) error {
	repo.guard.Lock()
	defer repo.guard.Unlock()

	if !repo.IsValidUserID(userID) {
		return model.ErrUserNotFound
	}

	for _, linkID := range links {
		if int(linkID) >= len(repo.Items) {
			continue
		}
		it := repo.Items[linkID]
		_, ok := it.Users[userID]
		if ok {
			it.Users[userID] = true
		}
	}
	return nil
}

func (repo *inMemoryRepo) IsValidUserID(id model.UserID) bool {
	return id.IsValid() && id < repo.NextUserID
}

func (repo *inMemoryRepo) saveOriginalURL(userID model.UserID, originalURL string) (model.LinkID, error) {
	if !repo.IsValidUserID(userID) {
		return 0, model.ErrUserNotFound
	}

	var err error

	idx := slices.IndexFunc(repo.Items, func(i *item) bool { return i.OriginalURL == originalURL })
	if idx == -1 {
		repo.addItem(originalURL)
		idx = len(repo.Items) - 1
	} else {
		err = model.ErrLinkAlreadyExists
	}

	it := repo.Items[idx]
	it.Users[userID] = false
	return model.LinkID(idx), err
}

func (repo *inMemoryRepo) Ping() error {
	return nil
}

func (repo *inMemoryRepo) Close() error {
	return nil
}

func (repo *inMemoryRepo) addItem(url string) {
	i := &item{
		OriginalURL: url,
		Users:       make(map[model.UserID]linkDeleted),
	}
	repo.Items = append(repo.Items, i)
}
