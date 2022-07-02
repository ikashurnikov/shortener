package storage

import (
	"fmt"
	"net/url"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func testStorage(newStorage func() Storage, t *testing.T) {
	testAddInvalidLongLink(newStorage(), t)
	testAddLongLink(newStorage(), t)
	testGetLongURL(newStorage(), t)
	testGetUserURLs(newStorage(), t)
}

func testAddInvalidLongLink(storage Storage, t *testing.T) {
	userID := InvalidUserID
	_, err := storage.AddLongURL(&userID, "", url.URL{})
	require.Error(t, err)
	require.Equal(t, InvalidUserID, userID)

	_, err = storage.AddLongURL(&userID, "ya", url.URL{})
	require.Error(t, err)
	require.Equal(t, InvalidUserID, userID)
}

func testAddLongLink(storage Storage, t *testing.T) {
	defer storage.Close()

	userID := InvalidUserID

	shortURL, err := storage.AddLongURL(&userID, "https://yandex.ru", url.URL{})
	require.NoError(t, err)
	require.True(t, shortURL != "")
	require.NotEqual(t, InvalidUserID, userID)

	// Добавялем туже самую ссылку
	shortURL2, err := storage.AddLongURL(&userID, "https://yandex.ru", url.URL{})
	require.NoError(t, err)
	require.Equal(t, shortURL, shortURL2)
	require.NotEqual(t, InvalidUserID, userID)

	// Новая ссыла
	shortURL3, err := storage.AddLongURL(&userID, "https://google.com", url.URL{})
	require.NoError(t, err)
	require.True(t, shortURL3 != "")
	require.NotEqual(t, shortURL3, shortURL)
	require.NotEqual(t, InvalidUserID, userID)
}

func testGetLongURL(storage Storage, t *testing.T) {
	defer storage.Close()

	_, err := storage.GetLongURL("yyyy")
	require.Error(t, err)

	userID := InvalidUserID

	for i := 0; i < 10; i++ {
		longURL := fmt.Sprintf("https://yandex.ru/%d", i)
		shortURL, err := storage.AddLongURL(&userID, longURL, url.URL{})
		require.NoError(t, err)

		longURL2, err := storage.GetLongURL(shortURL)
		require.NoError(t, err)
		require.Equal(t, longURL2, longURL)
	}
}

func testGetUserURLs(storage Storage, t *testing.T) {
	defer storage.Close()

	_, err := storage.GetUserURLs(0, url.URL{})
	require.Error(t, err)

	user1 := newTestUser()
	user2 := newTestUser()

	user1.addLongURL(storage, "http://share_url.ru", t)
	user2.addLongURL(storage, "http://share_url.ru", t)

	for i := 0; i < 4; i++ {
		user1.addLongURL(storage, fmt.Sprintf("https://user_1/%v", i), t)
		user1.addLongURL(storage, fmt.Sprintf("https://user_2/%v", i), t)
	}

	urls, err := storage.GetUserURLs(user1.id, url.URL{})
	require.NoError(t, err)
	require.True(t, user1.equal(urls))

	urls, err = storage.GetUserURLs(user2.id, url.URL{})
	require.NoError(t, err)
	require.True(t, user2.equal(urls))
}

type testUser struct {
	id   UserID
	urls []URLInfo
}

func newTestUser() testUser {
	return testUser{
		id: InvalidUserID,
	}
}

func (u *testUser) addLongURL(s Storage, longURL string, t *testing.T) {
	oldUserID := u.id

	shortURL, err := s.AddLongURL(&u.id, longURL, url.URL{})
	require.NoError(t, err)
	require.NotEqual(t, "", shortURL)

	if oldUserID != InvalidUserID {
		require.Equal(t, oldUserID, u.id)
	}

	for _, uinfo := range u.urls {
		if uinfo.LongURL == longURL {
			return
		}
	}

	u.urls = append(u.urls, URLInfo{LongURL: longURL, ShortURL: shortURL})
}

func (u *testUser) equal(urls []URLInfo) bool {
	sort.SliceStable(u.urls, func(i, j int) bool {
		return u.urls[i].LongURL < u.urls[j].LongURL
	})
	sort.SliceStable(urls, func(i, j int) bool {
		return urls[i].LongURL < urls[j].LongURL
	})
	return reflect.DeepEqual(urls, u.urls)
}
