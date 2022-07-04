package storage

import (
	"github.com/vmihailenco/msgpack/v5"
	"sync"
)

type (
	InMemoryStorage struct {
		store linksStore
		guard sync.RWMutex
	}

	linksStore struct {
		Links     linksMap   `msgpack:"links"`
		UsersData []userData `msgpack:"users_data"`
	}

	linksMap struct {
		LinkToID   map[string]LinkID `msgpack:"link_to_id"`
		IDToLink   map[LinkID]string `msgpack:"id_to_link"`
		NextLinkID LinkID            `msgpack:"nextId"`
	}

	userData struct {
		// Список всех ссылок пользователя
		Links map[string]LinkID `msgpack:"links"`
	}
)

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		store: newLinksStore(),
	}
}

func (s *InMemoryStorage) MarshalMsgpack() ([]byte, error) {
	s.guard.RLock()
	defer s.guard.RUnlock()

	return msgpack.Marshal(s.store)
}

func (s *InMemoryStorage) UnmarshalMsgpack(b []byte) error {
	st := newLinksStore()
	if err := msgpack.Unmarshal(b, &st); err != nil {
		return err
	}

	s.guard.Lock()
	defer s.guard.Unlock()

	s.store = st
	return nil
}

func (s *InMemoryStorage) InsertUser() (UserID, error) {
	s.guard.Lock()
	defer s.guard.Unlock()
	return s.store.insertUser(), nil
}

func (s *InMemoryStorage) InsertLink(userID UserID, link string) (LinkID, error) {
	s.guard.Lock()
	defer s.guard.Unlock()

	return s.store.insertLink(userID, link)
}

func (s *InMemoryStorage) InsertLinks(userID UserID, links []string) ([]LinkID, error) {
	s.guard.Lock()
	defer s.guard.Unlock()

	return s.store.insertLinks(userID, links)
}

func (s *InMemoryStorage) SelectLink(id LinkID) (string, error) {
	s.guard.RLock()
	defer s.guard.RUnlock()

	return s.store.selectLink(id)
}

func (s *InMemoryStorage) SelectUserLinks(id UserID) (map[string]LinkID, error) {
	s.guard.RLock()
	defer s.guard.RUnlock()

	return s.store.selectUserLink(id)
}

func (s *InMemoryStorage) Ping() error {
	return nil
}

func (s *InMemoryStorage) Close() error {
	return nil
}

// linkStore

func newLinksStore() linksStore {
	return linksStore{
		Links: linksMap{
			IDToLink:   make(map[LinkID]string),
			LinkToID:   make(map[string]LinkID),
			NextLinkID: 0,
		},
	}
}

func newUserData() userData {
	return userData{
		Links: make(map[string]LinkID),
	}
}

func (s *linksStore) insertUser() UserID {
	s.UsersData = append(s.UsersData, newUserData())
	idx := len(s.UsersData) - 1
	return UserID(idx)
}

func (s *linksStore) userData(id UserID) *userData {
	idx := int(id)
	if idx >= 0 && idx < len(s.UsersData) {
		return &s.UsersData[idx]
	}
	return nil
}

func (s *linksStore) insertLink(userID UserID, link string) (LinkID, error) {
	udata := s.userData(userID)
	if udata == nil {
		return 0, ErrUserNotFound
	}

	id, ok := s.findLinkID(link)
	if !ok {
		id = s.Links.NextLinkID
		s.Links.NextLinkID++
		s.Links.LinkToID[link] = id
		s.Links.IDToLink[id] = link
	}

	udata.Links[link] = id
	return id, nil
}

func (s *linksStore) insertLinks(userID UserID, urls []string) ([]LinkID, error) {
	res := make([]LinkID, len(urls))

	for i, url := range urls {
		id, err := s.insertLink(userID, url)
		if err != nil {
			return nil, err
		}
		res[i] = id
	}
	return res, nil
}

func (s *linksStore) selectLink(id LinkID) (string, error) {
	link, ok := s.findLink(id)
	if !ok {
		return "", ErrLinkNotFound
	}
	return link, nil
}

func (s *linksStore) selectUserLink(id UserID) (map[string]LinkID, error) {
	udata := s.userData(id)
	if udata == nil {
		return nil, ErrUserNotFound
	}

	res := make(map[string]LinkID)
	for k, v := range udata.Links {
		res[k] = v
	}
	return res, nil
}
func (s *linksStore) findLink(id LinkID) (string, bool) {
	url, ok := s.Links.IDToLink[id]
	return url, ok
}

func (s *linksStore) findLinkID(link string) (LinkID, bool) {
	id, ok := s.Links.LinkToID[link]
	return id, ok
}
