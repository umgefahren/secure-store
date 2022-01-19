package storage

import (
	"errors"
	"fmt"
	"io"
)

type Storage interface {
	NewBucket(bucket string) error
	Write(bucket, key string, data io.Reader) error
	Read(bucket, key string) (io.Reader, error)
	Delete(bucket, key string) error
	DeleteBucket(bucket string) error
	ListBuckets() ([]string, error)
}

func BucketDoesNotExist(id string) error {
	return errors.New(fmt.Sprintf("Bucket with id %v does not exist.", id))
}

func BucketAlreadyExists(id string) error {
	return errors.New(fmt.Sprintf("Bucket with id %v already exists.", id))
}

func ObjectDoesNotExists(key string) error {
	return errors.New(fmt.Sprintf("Object with key %v does not exist.", key))
}

func ObjectAlreadyExists(key string) error {
	return errors.New(fmt.Sprintf("Object with key %v already exists.", key))
}
