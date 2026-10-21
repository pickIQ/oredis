package store

import (
	"sync"
	"time"
)

type Store interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (value string, found bool)
	Delete(key string) (result bool)
}

type entry struct {
	value   string
	expTime time.Time
}

type InMemoryStore struct {
	m  map[string]entry
	mx sync.RWMutex
}

func isExpired(e entry) bool { return !e.expTime.IsZero() && time.Now().After(e.expTime) }

func (s *InMemoryStore) Set(key, value string, ttl time.Duration) error {
	var expTime time.Time

	if ttl != 0 {
		expTime = time.Now().Add(ttl)
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	s.m[key] = entry{
		value:   value,
		expTime: expTime,
	}
	return nil
}

func (s *InMemoryStore) Get(key string) (value string, found bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	e, ok := s.m[key]

	if !ok {
		return "", false
	}
	if isExpired(e) {
		return "", false
	}
	return e.value, true
}

func (s *InMemoryStore) Delete(key string) (result bool) {
	s.mx.Lock()
	defer s.mx.Unlock()
	e, ok := s.m[key]

	if !ok {
		return false
	}
	if isExpired(e) {
		return false
	}
	delete(s.m, key)
	return true
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		m: make(map[string]entry),
	}
}
