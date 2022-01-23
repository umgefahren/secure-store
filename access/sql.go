package access

import (
	"context"
	"crypto/sha512"
	"errors"
	"gorm.io/gorm"
	"log"
	"sync"
	"time"
)

type SQLStore struct {
	m          sync.Mutex
	db         *gorm.DB
	killer     chan<- *AKey
	timeKiller chan<- *AKey
}

type DBAKey struct {
	gorm.Model
	Expires         bool
	Ttl             *time.Time
	Limited         bool
	Limit           uint64
	UsedTimes       uint64
	BucketId        string
	KeyId           string
	UrlKey          string
	NeedsKey        bool
	ValidKeysID     uint
	ValidKeys       ValidKeys
	ResolveByUrlKey bool
}

type ValidKeys struct {
	gorm.Model
	keys [][sha512.Size]byte
}

func DBAKeyFromAKey(key *AKey) *DBAKey {
	ret := new(DBAKey)
	ret.Expires = key.expires
	ret.Ttl = key.ttl
	ret.Limited = key.limited
	ret.Limit = key.limit
	ret.UsedTimes = *key.usedTimes
	ret.BucketId = key.BucketId
	ret.KeyId = key.KeyId
	ret.UrlKey = key.UrlKey
	ret.NeedsKey = key.NeedsKey

	keys := make([][sha512.Size]byte, 0)
	if ret.NeedsKey {
		for array := range key.validKeys.set {
			keys = append(keys, array)
		}
	}

	ret.ValidKeys = ValidKeys{keys: keys}
	ret.ValidKeysID = ret.ValidKeys.ID
	ret.ResolveByUrlKey = key.ResolveByUrlKey
	return ret
}

func AKeyFromDBAKey(dbaKey *DBAKey) *AKey {
	ret := new(AKey)
	ret.expires = dbaKey.Expires
	ret.ttl = dbaKey.Ttl
	ret.limited = dbaKey.Limited
	ret.limit = dbaKey.Limit
	ret.usedTimes = &dbaKey.UsedTimes
	ret.BucketId = dbaKey.BucketId
	ret.KeyId = dbaKey.KeyId
	ret.UrlKey = dbaKey.UrlKey
	ret.NeedsKey = dbaKey.NeedsKey

	keys := new(HashSet)
	for _, key := range dbaKey.ValidKeys.keys {
		keys.Insert(key)
	}
	ret.validKeys = keys
	ret.ResolveByUrlKey = dbaKey.ResolveByUrlKey
	return ret
}

func NewSQLStore(genericDb gorm.Dialector) (*SQLStore, error) {
	db, err := gorm.Open(genericDb, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&DBAKey{})
	if err != nil {
		return nil, err
	}

	killerChan := make(chan *AKey)
	ret := new(SQLStore)
	ret.db = db
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

	return ret, nil
}

func (s *SQLStore) AddKey(key *AKey) error {
	s.m.Lock()
	defer func() {
		s.m.Unlock()
	}()
	var dbakey DBAKey
	result := s.db.First(&dbakey, "url_key = ?", key.UrlKey)
	log.Println(result.Error)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		if result.RowsAffected > 0 {
			return KeyAlreadyExists(key.UrlKey)
		}
	}
	dbaKey := DBAKeyFromAKey(key)
	result = s.db.Create(dbaKey)
	if result.Error != nil {
		return result.Error
	}
	if key.expires {
		s.timeKiller <- key
	}
	return nil
}

func (s *SQLStore) Access(ctx context.Context, urlKey string) (chan<- bool, *AKey, error) {
	s.m.Lock()
	defer func() {
		s.m.Unlock()
	}()
	var dbaKey DBAKey
	result := s.db.First(&dbaKey, "url_key = ?", urlKey)
	if result.RowsAffected == 0 {
		return nil, nil, KeyDoesntExist(urlKey)
	}
	id := dbaKey.ID
	backChan := make(chan bool)
	go func() {
		select {
		case <-ctx.Done():
			return
		case resp := <-backChan:
			s.m.Lock()
			defer func() {
				s.m.Unlock()
			}()
			var dbKey DBAKey
			s.db.First(&dbKey, id)
			if resp {
				s.db.Model(&DBAKey{}).Where("url_key = ?", urlKey).Update("used_times", dbKey.UsedTimes+1)
			}
			if dbKey.UsedTimes >= dbKey.Limit && dbKey.Limited {
				s.killer <- AKeyFromDBAKey(&dbKey)
			}
			return
		}
	}()
	return backChan, AKeyFromDBAKey(&dbaKey), nil
}

func (s *SQLStore) DeleteKey(key *AKey) error {
	s.m.Lock()
	defer func() {
		s.m.Unlock()
	}()
	var dbaKey DBAKey
	result := s.db.Where("url_key = ?", key.UrlKey).Delete(&dbaKey)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
