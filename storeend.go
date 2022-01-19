package main

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
	"secure-store/metadata"
	"secure-store/security"
	"secure-store/storage"
)

func CreateStreamReader(key []byte, data io.Reader) *cipher.StreamReader {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}

	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])
	reader := &cipher.StreamReader{
		S: stream,
		R: data,
	}
	return reader
}

type CompoundStore struct {
	metadata metadata.MetadataStore
	security security.SecurityStore
	storage  storage.Storage
}

func (c *CompoundStore) NewBucket(bucket string) error {
	err := c.metadata.NewBucket(bucket)
	if err != nil {
		return err
	}
	err = c.security.NewBucket(bucket)
	if err != nil {
		return err
	}
	err = c.storage.NewBucket(bucket)
	if err != nil {
		return err
	}
	return nil
}

func (c *CompoundStore) Write(bucketId, keyId string, metadata *metadata.Metadata, key security.EncryptionKey, data io.Reader) error {
	err := c.metadata.Write(bucketId, keyId, metadata)
	if err != nil {
		return err
	}
	err = c.security.WriteKey(bucketId, keyId, key)
	if err != nil {
		return err
	}

	reader := CreateStreamReader(key.Key, data)

	err = c.storage.Write(bucketId, keyId, reader)
	if err != nil {
		return err
	}
	return nil
}

func (c *CompoundStore) Read(bucketId, keyId string) (*metadata.Metadata, io.Reader, error) {
	meta, err := c.metadata.Read(bucketId, keyId)
	if err != nil {
		return nil, nil, err
	}
	key, err := c.security.ReadKey(bucketId, keyId)
	if err != nil {
		return nil, nil, err
	}
	data, err := c.storage.Read(bucketId, keyId)
	if err != nil {
		return nil, nil, err
	}
	reader := CreateStreamReader(key.Key, data)
	return meta, reader, nil
}

func (c *CompoundStore) Delete(bucketId, keyId string) error {
	err := c.metadata.Delete(bucketId, keyId)
	if err != nil {
		return err
	}
	err = c.security.DeleteKey(bucketId, keyId)
	if err != nil {
		return err
	}
	err = c.storage.Delete(bucketId, keyId)
	if err != nil {
		return err
	}
	return nil
}

func (c *CompoundStore) DeleteBucket(bucket string) error {
	err := c.metadata.DeleteBucket(bucket)
	if err != nil {
		return err
	}
	err = c.security.DeleteBucket(bucket)
	if err != nil {
		return err
	}
	err = c.storage.DeleteBucket(bucket)
	if err != nil {
		return err
	}
	return nil
}

func (c *CompoundStore) ListBuckets() ([]string, error) {
	return c.metadata.ListBuckets()
}
