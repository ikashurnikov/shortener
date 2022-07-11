package repo

import "testing"

func TestInMemoryRepo(t *testing.T) {
	testStorage(func() Repo {
		return NewInMemoryRepo()
	}, t)
}
