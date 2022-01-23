package metadata

import "testing"

func TestBucketsMemory(t *testing.T) {
	db := NewMemoryStore()
	BucketTest(t, db)
}

func TestWriteAndReadMemory(t *testing.T) {
	db := NewMemoryStore()
	ReadAndWriteTest(t, db)
}

func TestDeleteMemory(t *testing.T) {
	db := NewMemoryStore()
	DeleteTest(t, db)
}
