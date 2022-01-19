package storage

import (
	"bufio"
	"bytes"
	"io"
	"sync"
)

type MemoryStorage struct {
	m sync.Mutex
	i map[string]map[string][]byte
}

func New() *MemoryStorage {
	ret := new(MemoryStorage)
	ret.i = make(map[string]map[string][]byte)
	ret.m = sync.Mutex{}
	return ret
}

func (m *MemoryStorage) NewBucket(bucket string) error {
	m.m.Lock()
	defer m.m.Unlock()
	_, ok := m.i[bucket]
	if ok {
		return BucketAlreadyExists(bucket)
	}
	m.i[bucket] = make(map[string][]byte)
	return nil
}

func (m *MemoryStorage) Write(bucket, key string, data io.Reader) error {
	m.m.Lock()
	defer m.m.Unlock()
	if bucketMap, ok := m.i[bucket]; ok {
		_, ok := bucketMap[key]
		if ok {
			return ObjectAlreadyExists(key)
		}
		internalData, err := io.ReadAll(data)
		if err != nil {
			return err
		}
		bucketMap[key] = internalData
	} else {
		return BucketDoesNotExist(bucket)
	}
	return nil
}

func (m *MemoryStorage) Read(bucket, key string) (io.Reader, error) {
	m.m.Lock()
	defer m.m.Unlock()
	if bucketMap, ok := m.i[bucket]; ok {
		mapData, ok := bucketMap[key]
		if !ok {
			return nil, ObjectDoesNotExists(key)
		}
		data := make([]byte, len(mapData))
		copy(data, mapData)
		reader := bytes.NewReader(data)
		bufReader := bufio.NewReader(reader)
		return bufReader, nil
	}
	return nil, BucketDoesNotExist(bucket)
}

func (m *MemoryStorage) Delete(bucket, key string) error {
	m.m.Lock()
	defer m.m.Unlock()
	if bucketMap, ok := m.i[bucket]; ok {
		_, ok := bucketMap[key]
		if !ok {
			return ObjectDoesNotExists(key)
		}
		delete(bucketMap, key)
		return nil
	}
	return BucketDoesNotExist(bucket)
}

func (m *MemoryStorage) DeleteBucket(bucket string) error {
	m.m.Lock()
	defer m.m.Unlock()
	_, exists := m.i[bucket]
	if !exists {
		return BucketDoesNotExist(bucket)
	}
	delete(m.i, bucket)
	return nil
}

func (m *MemoryStorage) ListBuckets() ([]string, error) {
	m.m.Lock()
	defer m.m.Unlock()
	ret := make([]string, 0)
	for s, _ := range m.i {
		ret = append(ret, s)
	}
	return ret, nil
}
