package security

import (
	"secure-store/storage"
	"sync"
)

type MemorySecurityStore struct {
	m sync.Mutex
	i map[string]map[string]EncryptionKey
}

func NewMemorySecurityStore() *MemorySecurityStore {
	ret := new(MemorySecurityStore)
	ret.m = sync.Mutex{}
	ret.i = make(map[string]map[string]EncryptionKey)
	return ret
}

func (m *MemorySecurityStore) NewBucket(bucket string) error {
	m.m.Lock()
	defer m.m.Unlock()
	_, ok := m.i[bucket]
	if ok {
		return storage.BucketAlreadyExists(bucket)
	}
	m.i[bucket] = make(map[string]EncryptionKey)
	return nil
}

func (m *MemorySecurityStore) WriteKey(bucketId, keyId string, key EncryptionKey) error {
	m.m.Lock()
	defer m.m.Unlock()
	bucket, ok := m.i[bucketId]
	if !ok {
		return storage.BucketDoesNotExist(bucketId)
	}
	_, ok = bucket[keyId]
	if ok {
		return storage.ObjectAlreadyExists(keyId)
	}
	bucket[keyId] = key
	return nil
}

func (m *MemorySecurityStore) ReadKey(bucketId, keyId string) (EncryptionKey, error) {
	m.m.Lock()
	defer m.m.Unlock()
	bucket, ok := m.i[bucketId]
	if !ok {
		return EncryptionKey{}, storage.BucketDoesNotExist(bucketId)
	}
	key, ok := bucket[keyId]
	if !ok {
		return EncryptionKey{}, storage.ObjectDoesNotExists(keyId)
	}
	return key, nil
}

func (m *MemorySecurityStore) DeleteKey(bucketId, keyId string) error {
	m.m.Lock()
	defer m.m.Unlock()
	bucket, ok := m.i[bucketId]
	if !ok {
		return storage.BucketDoesNotExist(bucketId)
	}
	_, ok = bucket[keyId]
	if !ok {
		return storage.ObjectDoesNotExists(keyId)
	}
	delete(bucket, keyId)
	return nil
}

func (m *MemorySecurityStore) DeleteBucket(bucket string) error {
	m.m.Lock()
	defer m.m.Unlock()
	_, exists := m.i[bucket]
	if !exists {
		return storage.BucketDoesNotExist(bucket)
	}
	delete(m.i, bucket)
	return nil
}
