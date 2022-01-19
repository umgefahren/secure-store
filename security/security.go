package security

import "crypto/rand"

type SecurityStore interface {
	NewBucket(bucket string) error
	WriteKey(bucketId, keyId string, key EncryptionKey) error
	ReadKey(bucketId, keyId string) (EncryptionKey, error)
	DeleteKey(bucketId, keyId string) error
	DeleteBucket(bucket string) error
}

type EncryptionKey struct {
	Key []byte
}

func NewEncryptionKey() EncryptionKey {
	keyData := make([]byte, 32)
	l, err := rand.Read(keyData)
	if err != nil {
		panic(err)
	}
	if len(keyData) != l {
		panic("Fuck no")
	}
	ret := new(EncryptionKey)
	ret.Key = keyData
	r := *ret
	return r
}
