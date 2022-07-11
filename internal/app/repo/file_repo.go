package repo

import (
	"errors"
	"os"
	"sync"

	"github.com/ikashurnikov/shortener/internal/app/model"
)

type fileRepo struct {
	filename string
	cache    *inMemoryRepo
	guard    sync.Mutex
}

func NewFileRepo(filename string) (*fileRepo, error) {
	repo := &fileRepo{
		filename: filename,
		cache:    NewInMemoryRepo(),
	}

	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (repo *fileRepo) AddUser() (model.UserID, error) {
	userID, err := repo.cache.AddUser()
	if err == nil {
		err = repo.save()
	}
	return userID, err
}

func (repo *fileRepo) SaveOriginalURL(userID model.UserID, originalURL string) (model.LinkID, error) {
	linkID, err := repo.cache.SaveOriginalURL(userID, originalURL)
	if err == nil {
		err = repo.save()
	}
	return linkID, err
}

func (repo *fileRepo) SaveOriginalURLs(userID model.UserID, originalURLs []string) ([]model.LinkID, error) {
	linkIDs, err := repo.cache.SaveOriginalURLs(userID, originalURLs)
	if err == nil {
		err = repo.save()
	}
	return linkIDs, err
}

func (repo *fileRepo) GetOriginalURLByID(id model.LinkID) (string, error) {
	return repo.cache.GetOriginalURLByID(id)
}

func (repo *fileRepo) GetOriginalURLsByUserID(id model.UserID) (map[string]model.LinkID, error) {
	return repo.cache.GetOriginalURLsByUserID(id)
}

func (repo *fileRepo) Ping() error {
	return nil
}

func (repo *fileRepo) Close() error {
	return nil
}

func (repo *fileRepo) save() error {
	repo.guard.Lock()
	defer repo.guard.Unlock()

	file, err := os.OpenFile(repo.filename, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}

	if err = repo.cache.Serialize(file); err != nil {
		_ = file.Close()
		return err
	}

	return file.Close()
}

func (repo *fileRepo) load() error {
	stat, err := os.Stat(repo.filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if stat.Size() == 0 {
		return nil
	}

	file, err := os.OpenFile(repo.filename, os.O_RDONLY, 0664)
	if err != nil {
		return err
	}

	return repo.cache.Deserialize(file)
}
