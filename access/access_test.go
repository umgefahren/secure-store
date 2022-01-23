package access

import (
	"context"
	"sync"
	"testing"
)

var Buckets = []string{"Bucket1", "Bucket2", "Bucket3", "Bucket4"}
var Keys = []string{"Key1", "Key2", "Key3", "Key4"}
var UrlKeys = []string{"UrlKey1", "UrlKey2", "UrlKey3", "UrlKey4"}

var Options = []KeyOptions{
	{
		expires:         false,
		ttl:             nil,
		limited:         false,
		limit:           100,
		needsKey:        false,
		validKeys:       nil,
		resolveByUrlKey: false,
	},
	{
		expires:         false,
		ttl:             nil,
		limited:         false,
		limit:           200,
		needsKey:        false,
		validKeys:       nil,
		resolveByUrlKey: false,
	},
	{
		expires:         false,
		ttl:             nil,
		limited:         false,
		limit:           300,
		needsKey:        false,
		validKeys:       nil,
		resolveByUrlKey: false,
	},
	{
		expires:         false,
		ttl:             nil,
		limited:         false,
		limit:           400,
		needsKey:        false,
		validKeys:       nil,
		resolveByUrlKey: false,
	},
}

func AddAndDeleteTest(t *testing.T, db AccessStore) {
	keys := make([]*AKey, 0)
	var wg sync.WaitGroup

	for idx, urlKey := range UrlKeys {
		bucket := Buckets[idx]
		key := Keys[idx]
		aKey := NewAccessKey(bucket, key, urlKey, Options[idx])
		keys = append(keys, aKey)
		wg.Add(1)
		go func() {
			defer wg.Done()
			errGo := db.AddKey(aKey)
			if errGo != nil {
				t.Errorf("Adding key %v -> %v -> %v", aKey.BucketId, aKey.KeyId, aKey.UrlKey)
				t.Error(errGo)
				t.Fail()
				return
			}
		}()
	}
	wg.Wait()
	for _, aKey := range keys {
		aKey := aKey
		wg.Add(1)
		go func() {
			defer wg.Done()
			errGo := db.DeleteKey(aKey)
			if errGo != nil {
				t.Errorf("Deleted key %v -> %v -> %v", aKey.BucketId, aKey.KeyId, aKey.UrlKey)
				t.Error(errGo)
				t.Fail()
				return
			}
		}()
	}
}

func AddAndAccessTest(t *testing.T, db AccessStore) {
	keys := make([]*AKey, 0)
	var wg sync.WaitGroup

	for idx, urlKey := range UrlKeys {
		bucket := Buckets[idx]
		key := Keys[idx]
		aKey := NewAccessKey(bucket, key, urlKey, Options[idx])
		keys = append(keys, aKey)
		wg.Add(1)
		go func() {
			defer wg.Done()
			errGo := db.AddKey(aKey)
			if errGo != nil {
				t.Errorf("Adding key %v -> %v -> %v", aKey.BucketId, aKey.KeyId, aKey.UrlKey)
				t.Error(errGo)
				t.Fail()
				return
			}
		}()
	}
	wg.Wait()
	for _, aKey := range keys {
		aKey := aKey
		wg.Add(1)
		go func() {
			defer wg.Done()
			channel, bKey, errGo := db.Access(context.TODO(), aKey.UrlKey)
			if errGo != nil {
				t.Errorf("Access key %v -> %v -> %v", aKey.BucketId, aKey.KeyId, aKey.UrlKey)
				t.Error(errGo)
				t.Fail()
				return
			}
			defer func() {
				channel <- true
			}()
			if bKey.UrlKey != aKey.UrlKey {
				t.Errorf("Returned wrong access key")
				t.Fail()
				return
			}
		}()
	}
	wg.Wait()
}
