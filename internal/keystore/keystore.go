package keystore

import (
	"sync"

	"github.com/vitalvas/oneauth/internal/agentkey"
)

type Store struct {
	keys map[string]*agentkey.Key
	lock sync.Mutex
}

func New() *Store {
	return &Store{
		keys: make(map[string]*agentkey.Key),
	}
}

func (s *Store) Len() int {
	s.lock.Lock()
	defer s.lock.Unlock()

	return len(s.keys)
}

func (s *Store) List() []*agentkey.Key {
	s.lock.Lock()
	defer s.lock.Unlock()

	keys := make([]*agentkey.Key, 0, len(s.keys))
	for _, key := range s.keys {
		keys = append(keys, key)
	}

	return keys
}

func (s *Store) Get(fp string) (*agentkey.Key, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if key, ok := s.keys[fp]; ok {
		return key, true
	}

	return nil, false
}

func (s *Store) Add(key *agentkey.Key) bool {
	if _, ok := s.Get(key.Fingerprint()); ok {
		return false
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.keys[key.Fingerprint()] = key

	return true
}

func (s *Store) Remove(fp string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.keys[fp]; ok {
		delete(s.keys, fp)
		return true
	}

	return false
}

func (s *Store) RemoveAll() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.keys = make(map[string]*agentkey.Key)
}
