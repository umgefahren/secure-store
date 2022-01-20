package storage

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

const secureStoreJsonName = ".secure-store.json"

type hashSetMap map[string]*interface{}

type hashSet struct {
	set hashSetMap
}

func (m *hashSet) Insert(obj string) {
	m.set[obj] = nil
}

func (m *hashSet) Contains(obj string) bool {
	_, ok := m.set[obj]
	return ok
}

func (m *hashSet) Delete(obj string) {
	delete(m.set, obj)
}

func newHashSet() *hashSet {
	set := make(hashSetMap)
	ret := new(hashSet)
	ret.set = set
	return ret
}

type FsStorage struct {
	m        sync.Mutex
	rootPath string
	memRep   map[string]*hashSet
}

func NewFsStorage(root string) (*FsStorage, error) {
	ret := new(FsStorage)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		err = os.MkdirAll(root, 0777)
		if err != nil {
			return nil, err
		}
	} else {

		fCounter := 0
		secureStoreFileExists := false
		entrys, err := os.ReadDir(root)
		if err != nil {
			log.Printf("Error while listing directory %v", err)
			return nil, err
		}
		for _, entry := range entrys {
			if entry.IsDir() {
				continue
			}
			if entry.Name() == secureStoreJsonName {
				secureStoreFileExists = true
				continue
			}
			fCounter += 1
		}
		if fCounter > 0 && !secureStoreFileExists {
			return nil, errors.New("given directory contains files but no config file for secure store")
		}
	}
	file, err := os.Create(fmt.Sprintf("%v/%v", root, secureStoreJsonName))
	if err != nil {
		log.Printf("Error while creating .secure-store.json %v", err)
		return nil, err
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}
	ret.memRep = make(map[string]*hashSet)
	ret.rootPath = root
	ret.m = sync.Mutex{}
	return ret, nil
}

func (f *FsStorage) GetRootBucket(bucketName string) string {
	return fmt.Sprintf("%v/%v", f.rootPath, bucketName)
}

func (f *FsStorage) GetRootKey(bucketName, keyName string) string {
	return fmt.Sprintf("%v/%v", f.GetRootBucket(bucketName), keyName)
}

func (f *FsStorage) NewBucket(bucket string) error {
	f.m.Lock()
	defer f.m.Unlock()
	_, ok := f.memRep[bucket]
	if ok {
		return BucketAlreadyExists(bucket)
	}
	f.memRep[bucket] = newHashSet()
	bucketRootDir := f.GetRootBucket(bucket)
	log.Printf("Creating bucket dir @ %v", bucketRootDir)
	err := os.Mkdir(bucketRootDir, 0777)
	if err != nil {
		return err
	}
	return nil
}

func (f *FsStorage) Write(bucket, key string, data io.Reader) error {
	f.m.Lock()
	keySet, ok := f.memRep[bucket]
	if !ok {
		f.m.Unlock()
		return BucketDoesNotExist(bucket)
	}
	if keySet.Contains(key) {
		f.m.Unlock()
		return ObjectAlreadyExists(key)
	}
	keySet.Insert(key)
	f.m.Unlock()
	file, err := os.Create(f.GetRootKey(bucket, key))
	if err != nil {
		return err
	}
	proxyReader := bufio.NewReader(data)
	_, err = file.ReadFrom(proxyReader)
	if err != nil {
		return err
	}
	return nil
}

func (f *FsStorage) Read(bucket, key string) (io.Reader, error) {
	f.m.Lock()
	keySet, ok := f.memRep[bucket]
	if !ok {
		f.m.Unlock()
		return nil, BucketDoesNotExist(bucket)
	}
	if !keySet.Contains(key) {
		f.m.Unlock()
		return nil, ObjectDoesNotExists(key)
	}
	f.m.Unlock()
	file, err := os.Open(f.GetRootKey(bucket, key))
	if err != nil {
		return nil, err
	}
	proxyReader := bufio.NewReader(file)
	return proxyReader, nil
}

func (f *FsStorage) Delete(bucket, key string) error {
	f.m.Lock()
	keySet, ok := f.memRep[bucket]
	if !ok {
		f.m.Unlock()
		return BucketDoesNotExist(bucket)
	}
	if !keySet.Contains(key) {
		f.m.Unlock()
		return ObjectDoesNotExists(key)
	}
	f.memRep[bucket].Delete(key)
	f.m.Unlock()
	err := os.Remove(f.GetRootKey(bucket, key))
	if err != nil {
		return err
	}
	return nil
}

func (f *FsStorage) DeleteBucket(bucket string) error {
	f.m.Lock()
	defer f.m.Unlock()
	_, ok := f.memRep[bucket]
	if !ok {
		return BucketDoesNotExist(bucket)
	}
	err := os.RemoveAll(f.GetRootBucket(bucket))
	delete(f.memRep, bucket)
	return err
}

func (f *FsStorage) ListBuckets() ([]string, error) {
	f.m.Lock()
	defer f.m.Unlock()
	ret := make([]string, len(f.memRep))
	for bucketName := range f.memRep {
		ret = append(ret, bucketName)
	}
	return ret, nil
}
