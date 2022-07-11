package repo

import (
	"fmt"
	"github.com/ikashurnikov/shortener/internal/app/model"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func testStorage(newRepo func() Repo, t *testing.T) {
	testSaveOriginalURL(newRepo(), t)
	testGetOriginalURLByID(newRepo(), t)
	testGetOriginalURLsByUserID(newRepo(), t)
}

func testSaveOriginalURL(repo Repo, t *testing.T) {
	userID, err := repo.AddUser()
	require.NoError(t, err)

	id, err := repo.SaveOriginalURL(userID, "https://yandex.ru")
	require.NoError(t, err)

	// Добавялем туже самую ссылку
	id2, err := repo.SaveOriginalURL(userID, "https://yandex.ru")
	require.Error(t, err, model.ErrLinkAlreadyExists)
	require.Equal(t, id, id2)

	// Новая ссыла
	id3, err := repo.SaveOriginalURL(userID, "https://google.com")
	require.NoError(t, err)
	require.NotEqual(t, id2, id3)
}

func testGetOriginalURLByID(repo Repo, t *testing.T) {
	userID, err := repo.AddUser()
	require.NoError(t, err)

	_, err = repo.GetOriginalURLByID(0)
	require.Error(t, err)

	for i := 0; i < 10; i++ {
		origURL := fmt.Sprintf("https://yandex.ru/%d", i)
		id, err := repo.SaveOriginalURL(userID, origURL)
		require.NoError(t, err)

		longURL2, err := repo.GetOriginalURLByID(id)
		require.NoError(t, err)
		require.Equal(t, longURL2, origURL)
	}
}

func testGetOriginalURLsByUserID(repo Repo, t *testing.T) {
	_, err := repo.GetOriginalURLsByUserID(0)
	require.Error(t, err)

	user1 := newTestUser(repo, t)
	user2 := newTestUser(repo, t)

	user1.saveOriginalURL("http://share_url.ru")
	user2.saveOriginalURL("http://share_url.ru")

	for i := 0; i < 4; i++ {
		user1.saveOriginalURL(fmt.Sprintf("https://user_1/%v", i))
		user1.saveOriginalURL(fmt.Sprintf("https://user_2/%v", i))
	}

	links, err := repo.GetOriginalURLsByUserID(user1.id)
	require.NoError(t, err)
	require.True(t, user1.equal(links))

	links, err = repo.GetOriginalURLsByUserID(user2.id)
	require.NoError(t, err)
	require.True(t, user2.equal(links))
}

type testUser struct {
	id    model.UserID
	links map[string]model.LinkID
	repo  Repo
	t     *testing.T
}

func newTestUser(repo Repo, t *testing.T) testUser {
	id, err := repo.AddUser()
	require.NoError(t, err)

	return testUser{
		id:    id,
		links: make(map[string]model.LinkID),
		repo:  repo,
		t:     t}
}

func (u *testUser) saveOriginalURL(origURL string) {
	id, err := u.repo.SaveOriginalURL(u.id, origURL)
	if err != nil {
		require.Error(u.t, model.ErrLinkAlreadyExists)
	}
	u.links[origURL] = id
}

func (u *testUser) equal(links map[string]model.LinkID) bool {
	return reflect.DeepEqual(u.links, links)
}
