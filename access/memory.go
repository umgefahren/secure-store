package access

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type MemoryStore struct {
	m          sync.Map
	killer     chan<- *AKey
	timeKiller chan<- *AKey
}

func NewMemoryStore() *MemoryStore {
	ret := &MemoryStore{}
	killerChan := make(chan *AKey)
	ret.killer = killerChan
	go func() {
		for {
			k := <-killerChan
			if k == nil {
				return
			}
			_ = ret.DeleteKey(k)
		}
	}()
	timeKiller := make(chan *AKey)
	ret.timeKiller = timeKiller
	go func() {
		for {
			k := <-timeKiller
			if k == nil {
				return
			}
			duration := time.Until(*k.ttl)
			go func() {
				timer := time.NewTimer(duration)
				_ = <-timer.C
				killerChan <- k
			}()
		}
	}()
	return ret
}

func (m *MemoryStore) AddKey(key *AKey) error {
	_, ok := m.m.Load(key.UrlKey)
	if ok {
		return KeyAlreadyExists(key.UrlKey)
	}
	m.m.Store(key.UrlKey, key)
	if key.expires {
		m.timeKiller <- key
	}
	return nil
}

func (m *MemoryStore) Access(ctx context.Context, urlKey string) (chan<- bool, *AKey, error) {
	keyTmp, ok := m.m.Load(urlKey)
	if !ok {
		return nil, nil, KeyDoesntExist(urlKey)
	}
	key := keyTmp.(*AKey)
	backChan := make(chan bool)
	go func() {
		select {
		case <-ctx.Done():
			return
		case resp := <-backChan:
			if resp {
				counterPointer := key.usedTimes
				atomic.AddUint64(counterPointer, 1)
			}
			counter := atomic.LoadUint64(key.usedTimes)
			if counter >= key.limit && key.limited {
				m.killer <- key
			}
			return
		}
	}()
	return backChan, key, nil
}

func (m *MemoryStore) DeleteKey(key *AKey) error {
	urlKey := key.UrlKey
	_, ok := m.m.Load(urlKey)
	if !ok {
		return KeyDoesntExist(urlKey)
	}
	m.m.Delete(urlKey)
	return nil
}
