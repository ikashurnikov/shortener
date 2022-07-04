package storage

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func testStorage(newStorage func() Storage, t *testing.T) {
	testInsertLink(newStorage(), t)
	testSelectLink(newStorage(), t)
	testSelectUserLinks(newStorage(), t)
}

func testInsertLink(storage Storage, t *testing.T) {
	defer storage.Close()

	userID, err := storage.InsertUser()
	require.NoError(t, err)

	id, err := storage.InsertLink(userID, "https://yandex.ru")
	require.NoError(t, err)

	// Добавялем туже самую ссылку
	id2, err := storage.InsertLink(userID, "https://yandex.ru")
	require.Error(t, err, ErrLinkAlreadyExists)
	require.Equal(t, id, id2)

	// Новая ссыла
	id3, err := storage.InsertLink(userID, "https://google.com")
	require.NoError(t, err)
	require.NotEqual(t, id2, id3)
}

func testSelectLink(storage Storage, t *testing.T) {
	defer storage.Close()

	userID, err := storage.InsertUser()
	require.NoError(t, err)

	_, err = storage.SelectLink(0)
	require.Error(t, err)

	for i := 0; i < 10; i++ {
		longURL := fmt.Sprintf("https://yandex.ru/%d", i)
		id, err := storage.InsertLink(userID, longURL)
		require.NoError(t, err)

		longURL2, err := storage.SelectLink(id)
		require.NoError(t, err)
		require.Equal(t, longURL2, longURL)
	}
}

func testSelectUserLinks(storage Storage, t *testing.T) {
	defer storage.Close()

	_, err := storage.SelectUserLinks(0)
	require.Error(t, err)

	user1 := newTestUser(storage, t)
	user2 := newTestUser(storage, t)

	user1.addLink(storage, "http://share_url.ru", t)
	user2.addLink(storage, "http://share_url.ru", t)

	for i := 0; i < 4; i++ {
		user1.addLink(storage, fmt.Sprintf("https://user_1/%v", i), t)
		user1.addLink(storage, fmt.Sprintf("https://user_2/%v", i), t)
	}

	links, err := storage.SelectUserLinks(user1.id)
	require.NoError(t, err)
	require.True(t, user1.equal(links))

	links, err = storage.SelectUserLinks(user2.id)
	require.NoError(t, err)
	require.True(t, user2.equal(links))
}

type testUser struct {
	id    UserID
	links map[string]LinkID
}

func newTestUser(storage Storage, t *testing.T) testUser {
	id, err := storage.InsertUser()
	require.NoError(t, err)

	return testUser{id: id, links: make(map[string]LinkID)}
}

func (u *testUser) addLink(s Storage, link string, t *testing.T) {
	id, err := s.InsertLink(u.id, link)
	if err != nil {
		require.Error(t, ErrLinkAlreadyExists)
	}
	u.links[link] = id
}

func (u *testUser) equal(links map[string]LinkID) bool {
	return reflect.DeepEqual(u.links, links)
}
