package storage

import (
	"github.com/ikashurnikov/shortener/internal/app/urlencoder"
	"github.com/vmihailenco/msgpack/v5"
	"sync"
)

type (
	InMemoryStorage struct {
		store store
		guard sync.RWMutex
		baseStorage
	}

	store struct {
		URLs      urlMap     `msgpack:"urls"`
		UsersData []userData `msgpack:"usersData"`
	}

	urlMap struct {
		URLToID map[string]uint32 `msgpack:"urls2id"`
		IDToURL map[uint32]string `msgpack:"id2urls"`
		NextID  uint32            `msgpack:"nextId"`
	}

	userData struct {
		// Список всех ссылок пользователя
		URLs map[string]uint32 `msgpack:"urls"`
	}
)

func NewInMemoryStorage() *InMemoryStorage {
	s := new(InMemoryStorage)
	s.store = newStore()
	s.urlEncoder = urlencoder.NewZBase32Encoder()
	s.baseStorage.storageImpl = s
	return s
}

func (s *InMemoryStorage) MarshalMsgpack() ([]byte, error) {
	s.guard.RLock()
	defer s.guard.RUnlock()

	return msgpack.Marshal(s.store)
}

func (s *InMemoryStorage) UnmarshalMsgpack(b []byte) error {
	st := newStore()
	if err := msgpack.Unmarshal(b, &st); err != nil {
		return err
	}

	s.guard.Lock()
	defer s.guard.Unlock()

	s.store = st
	return nil
}

func (s *InMemoryStorage) insertLongURL(userID *UserID, url string) (uint32, error) {
	s.guard.Lock()
	defer s.guard.Unlock()

	id := s.store.addURL(userID, url)
	return id, nil
}

func (s *InMemoryStorage) selectLongURL(id uint32) (string, error) {
	s.guard.RLock()
	defer s.guard.RUnlock()

	url, ok := s.store.findURL(id)
	if !ok {
		return "", ErrDecodingShortURL
	}

	return url, nil
}

func (s *InMemoryStorage) getUserURLs(userID UserID, newURLInfo newURLInfoFunc) ([]URLInfo, error) {
	s.guard.RLock()
	defer s.guard.RUnlock()

	udata := s.store.userData(&userID, false)
	if udata == nil {
		return nil, ErrUserNotFound
	}

	res := make([]URLInfo, 0, len(udata.URLs))

	for url, id := range udata.URLs {
		urlInfo, err := newURLInfo(id, url)
		if err != nil {
			return nil, err
		}
		res = append(res, urlInfo)
	}

	return res, nil
}

func (s *InMemoryStorage) close() error {
	return nil
}

func newStore() store {
	return store{
		URLs: urlMap{
			IDToURL: make(map[uint32]string),
			URLToID: make(map[string]uint32),
			NextID:  0,
		},
	}
}

func newUserData() userData {
	return userData{
		URLs: make(map[string]uint32),
	}
}

func (s *store) findURL(id uint32) (string, bool) {
	url, ok := s.URLs.IDToURL[id]
	return url, ok
}

func (s *store) findID(url string) (uint32, bool) {
	id, ok := s.URLs.URLToID[url]
	return id, ok
}

func (s *store) addURL(userID *UserID, url string) uint32 {
	id, ok := s.findID(url)
	if !ok {
		id = s.URLs.NextID
		s.URLs.NextID++
		s.URLs.URLToID[url] = id
		s.URLs.IDToURL[id] = url
	}

	d := s.userData(userID, true)
	d.URLs[url] = id
	return id
}

func (s *store) userData(id *UserID, add bool) *userData {
	index := int(*id)
	if index < 0 || index >= len(s.UsersData) {
		if !add {
			return nil
		}
		index = len(s.UsersData)
		*id = UserID(index)
		s.UsersData = append(s.UsersData, newUserData())
	}
	return &s.UsersData[index]
}
