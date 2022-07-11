package repo

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestFileRepo(t *testing.T) {
	filenames := make([]string, 0)
	defer func() {
		for _, filename := range filenames {
			os.Remove(filename)
		}
	}()

	testStorage(func() Repo {
		filename := uuid.New().String()
		filenames = append(filenames, filename)
		storage, err := NewFileRepo(filename)
		require.NoError(t, err)
		return storage
	}, t)
}

func TestFileStorage_ReadWrite(t *testing.T) {
	filename := uuid.New().String()
	defer os.Remove(filename)

	repo, err := NewFileRepo(filename)
	require.NoError(t, err)

	user1 := newTestUser(repo, t)
	user2 := newTestUser(repo, t)

	user1.saveOriginalURL("http://share_url.ru")
	user2.saveOriginalURL("http://share_url.ru")

	for i := 0; i < 4; i++ {
		user1.saveOriginalURL(fmt.Sprintf("https://user_1/%v", i))
		user1.saveOriginalURL(fmt.Sprintf("https://user_2/%v", i))
	}

	// Загружаем данные с диска.
	repo, err = NewFileRepo(filename)
	require.NoError(t, err)

	urls, err := repo.GetOriginalURLsByUserID(user1.id)
	require.NoError(t, err)
	require.True(t, user1.equal(urls))

	urls, err = repo.GetOriginalURLsByUserID(user2.id)
	require.NoError(t, err)
	require.True(t, user2.equal(urls))
}
