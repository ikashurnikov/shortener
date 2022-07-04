package storage

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
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

	user1 := newTestUser(s, t)
	user2 := newTestUser(s, t)

	user1.addLink(s, "http://share_url.ru", t)
	user2.addLink(s, "http://share_url.ru", t)

	for i := 0; i < 4; i++ {
		user1.addLink(s, fmt.Sprintf("https://user_1/%v", i), t)
		user1.addLink(s, fmt.Sprintf("https://user_2/%v", i), t)
	}
	// Закрываем хранилище и скидывем данные на диск.
	require.NoError(t, s.Close())

	// Загружаем данные с диска.
	s, err = NewFileStorage(filename)
	require.NoError(t, err)

	urls, err := s.SelectUserLinks(user1.id)
	require.NoError(t, err)
	require.True(t, user1.equal(urls))

	urls, err = s.SelectUserLinks(user2.id)
	require.NoError(t, err)
	require.True(t, user2.equal(urls))
}
