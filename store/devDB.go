package store

import (
	"github.com/redis-go/redis"
	"sync"
	"time"
)

type DevStore struct {
	items map[string]*redis.Item
	redis *redis.Redis
}

// in development
func NewDevStore(r *redis.Redis) redis.ItemStore {
	return &DevStore{
		items: map[string]*redis.Item{},
		redis: r,
	}
}

func (s *DevStore) Redis() *redis.Redis {
	return s.redis
}

func (s *DevStore) Mu() *sync.RWMutex {
	return s.Redis().Mu()
}

func (s *DevStore) Set(key *string, i *redis.Item) {
	s.Mu().Lock()
	defer s.Mu().Unlock()
	s.items[*key] = i
}

func (s *DevStore) Get(key *string) *redis.Item {
	s.Mu().RLock()
	i, ok := s.items[*key]
	s.Mu().RUnlock()
	if ok {
		return i
	}
	return nil
}

func (s *DevStore) Delete(key *string) bool {
	if i := s.Get(key); i != nil {
		item := *i
		item.OnDelete(key, s.Redis())
		s.Mu().Lock()
		defer s.Mu().Unlock()
		delete(s.items, *key)
		return true
	}
	return false
}

func (s *DevStore) Exists(key *string) bool {
	s.Mu().RLock()
	defer s.Mu().RUnlock()
	_, ok := s.items[*key]
	return ok
}

func (s *DevStore) Expire(key *string) time.Time {

}

func (s *DevStore) CheckExpire(key *string) bool {

}

func (s *DevStore) AddItem(key *string, i redis.Item) {
	s.items[*key] = &i
}

func (s *DevStore) DeleteItem(key *string) bool {
	if i, ok := s.GetItem(key); ok && i.OnDelete(s) {
		s.Mu().Lock()
		defer s.Mu().Unlock()
		delete(s.items, *key)
		return true
	}
	return false
}

func (s *DevStore) ItemExists(key *string) bool {
	_, ok := s.items[*key]
	return ok
}

func (s *DevStore) GetItem(key *string) (Item, bool) {
	s.Mu().Lock()
	defer s.Mu().Unlock()
	i, ok := s.items[*key]
	return *i, ok
}

func (s *DevStore) CheckExpireAll() uint64 {
	var c uint64
	s.Mu().RLock()
	items := s.items
	s.Mu().RUnlock()
	for k, i := range items {
		item := *i
		if item.Expired() {
			if s.DeleteItem(&k) {
				c++
			}
		}
	}
	return c
}

// Returns:
//
// -1 if item did not existed
//
// 0 if item did not expired
//
// 1 if item expired and deleted
//
// 2 if item is expired but could not delete due to OnDelete() of item returned false
func (s *DevStore) CheckExpire(key *string) int8 {
	i, ok := s.GetItem(key)
	if !ok {
		return int8(-1)
	}

	if i.Expired() {
		if s.DeleteItem(key) {
			return int8(1)
		}
		return int8(2)
	}
	return int8(0)
}
