package metadata

import (
	"errors"
	"secure-store/storage"
	"sync"
)

type MemoryStore struct {
	m sync.Mutex
	i map[string]map[string]*Metadata
}

func NewMemoryStore() *MemoryStore {
	ret := new(MemoryStore)
	ret.m = sync.Mutex{}
	ret.i = make(map[string]map[string]*Metadata)
	return ret
}

func (m *MemoryStore) Write(bucketId, keyId string, metadata *Metadata) error {
	m.m.Lock()
	defer m.m.Unlock()
	bucket := m.i[bucketId]
	bucket[keyId] = metadata
	return nil
}

func (m *MemoryStore) Read(bucketId, keyId string) (*Metadata, error) {
	m.m.Lock()
	defer m.m.Unlock()
	if bucket, ok := m.i[bucketId]; ok {
		if meta, ok := bucket[keyId]; ok {
			return meta, nil
		}
	}
	return nil, errors.New("there is no bucket")
}

func (m *MemoryStore) NewBucket(bucket string) error {
	m.m.Lock()
	defer m.m.Unlock()
	_, ok := m.i[bucket]
	if ok {
		return storage.BucketAlreadyExists(bucket)
	}
	m.i[bucket] = make(map[string]*Metadata)
	return nil
}

func (m *MemoryStore) Delete(bucket, key string) error {
	m.m.Lock()
	defer m.m.Unlock()
	if bucketMap, ok := m.i[bucket]; ok {
		_, ok := bucketMap[key]
		if !ok {
			return storage.ObjectDoesNotExists(key)
		}
		delete(bucketMap, key)
		m.i[bucket] = bucketMap
		return nil
	}
	return storage.BucketDoesNotExist(bucket)
}

func (m *MemoryStore) DeleteBucket(bucket string) error {
	m.m.Lock()
	defer m.m.Unlock()
	_, exists := m.i[bucket]
	if !exists {
		return storage.BucketDoesNotExist(bucket)
	}
	delete(m.i, bucket)
	return nil
}

func (m *MemoryStore) ListBuckets() ([]string, error) {
	m.m.Lock()
	defer m.m.Unlock()
	ret := make([]string, 0)
	for s := range m.i {
		ret = append(ret, s)
	}
	return ret, nil
}
