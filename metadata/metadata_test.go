package metadata

import (
	"sort"
	"sync"
	"testing"
)

var Buckets = []string{"Bucket1", "Bucket2", "Bucket3", "Bucket4"}
var Keys = []string{"Key1", "Key2", "Key3", "Key4"}
var Metas = []Metadata{{
	Length:   100,
	Filename: "file1.txt",
}, {
	Length:   200,
	Filename: "file2.txt",
}, {
	Length:   300,
	Filename: "file3.txt",
}, {
	Length:   400,
	Filename: "file4.txt",
}}

func BucketTest(t *testing.T, db MetadataStore) {
	for _, bucket := range Buckets {
		err := db.NewBucket(bucket)
		if err != nil {
			t.Fatal(err)
		}
	}
	retBuckets, err := db.ListBuckets()
	sort.Strings(retBuckets)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	for idx, bucket := range Buckets {
		if retBuckets[idx] != bucket {
			t.Errorf("Ret Bucket %v != Bucket %v", retBuckets[idx], bucket)
			t.Fail()
		}
	}
}

func ReadAndWriteTest(t *testing.T, db MetadataStore) {
	var wg sync.WaitGroup
	for _, bucket := range Buckets {
		err := db.NewBucket(bucket)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		for idx, key := range Keys {
			wg.Add(1)
			key := key
			idx := idx
			bucket := bucket
			go func() {
				defer wg.Done()
				errGo := db.Write(bucket, key, &Metas[idx])
				if errGo != nil {
					t.Errorf("Error during write of %v -> %v", bucket, key)
					t.Error(errGo)
					t.Fail()
				}
			}()
		}
	}
	wg.Wait()

	var wg2 sync.WaitGroup
	for _, bucket := range Buckets {
		for idx, key := range Keys {
			wg2.Add(1)
			key := key
			idx := idx
			bucket := bucket
			go func() {
				defer wg2.Done()
				metaGo, errGo := db.Read(bucket, key)
				if errGo != nil {
					t.Errorf("Error during read of %v -> %v", bucket, key)
					t.Fail()
					return
				}
				if *metaGo != Metas[idx] {
					t.Errorf("Wrong meta extracted from %v -> %v", bucket, key)
					t.Fail()
				}
			}()
		}
	}
	wg2.Wait()
}

func DeleteTest(t *testing.T, db MetadataStore) {
	var wg sync.WaitGroup

	for _, bucket := range Buckets {
		err := db.NewBucket(bucket)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		for idx, key := range Keys {
			wg.Add(1)
			key := key
			idx := idx
			bucket := bucket
			go func() {
				defer wg.Done()
				errGo := db.Write(bucket, key, &Metas[idx])
				if errGo != nil {
					t.Errorf("Error during write of %v -> %v", bucket, key)
					t.Error(errGo)
					t.Fail()
				}
			}()
		}
	}
	wg.Wait()
	var wg2 sync.WaitGroup
	for idy, bucket := range Buckets {
		if idy != len(bucket)-2 {
			for _, key := range Keys {
				wg2.Add(1)
				bucket := bucket
				key := key
				go func() {
					defer wg2.Done()
					errGo := db.Delete(bucket, key)
					if errGo != nil {
						t.Errorf("Error during delete of %v -> %v", bucket, key)
						t.Error(errGo)
						t.Fail()
					}
				}()
			}
		} else {
			wg2.Add(1)
			bucket := bucket
			go func() {
				defer wg2.Done()
				errGo := db.DeleteBucket(bucket)
				if errGo != nil {
					t.Errorf("Error during deletion of bucket %v", bucket)
					t.Error(errGo)
					t.Fail()
				}
			}()
		}
	}
	wg2.Wait()
}
