package storage

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/url"
	"os"
	"testing"
)

func TestFileStorage(t *testing.T) {
	filenames := make([]string, 0)
	defer func() {
		for _, filename := range filenames {
			os.Remove(filename)
		}
	}()

	testStorage(func() Storage {
		filename := uuid.New().String()
		filenames = append(filenames, filename)
		storage, err := NewFileStorage(filename)
		require.NoError(t, err)
		return storage
	}, t)
}

func TestFileStorage_ReadWrite(t *testing.T) {
	filename := uuid.New().String()
	defer os.Remove(filename)

	s, err := NewFileStorage(filename)
	require.NoError(t, err)

	user1 := newTestUser()
	user2 := newTestUser()

	user1.addLongURL(s, "http://share_url.ru", t)
	user2.addLongURL(s, "http://share_url.ru", t)

	for i := 0; i < 4; i++ {
		user1.addLongURL(s, fmt.Sprintf("https://user_1/%v", i), t)
		user1.addLongURL(s, fmt.Sprintf("https://user_2/%v", i), t)
	}
	// Закрываем хранилище и скидывем данные на диск.
	require.NoError(t, s.Close())

	// Загружаем данные с диска.
	s, err = NewFileStorage(filename)
	require.NoError(t, err)

	urls, err := s.GetUserURLs(user1.id, url.URL{})
	require.NoError(t, err)
	require.True(t, user1.equal(urls))

	urls, err = s.GetUserURLs(user2.id, url.URL{})
	require.NoError(t, err)
	require.True(t, user2.equal(urls))
}
