package metadata

import (
	"errors"
	"gorm.io/gorm"
	"sync"
)

type SQLStore struct {
	db      *gorm.DB
	m       sync.Mutex
	buckets []string
}

type SQLMetadata struct {
	gorm.Model
	Length   int64
	Filename string
	BucketId string
	KeyId    string
}

func MetadataFromSQLMetadata(sqlMeta *SQLMetadata) *Metadata {
	return &Metadata{
		Length:   sqlMeta.Length,
		Filename: sqlMeta.Filename,
	}
}

func SQLMetadataFromMetadata(bucketId, keyId string, meta *Metadata) *SQLMetadata {
	return &SQLMetadata{
		Length:   meta.Length,
		Filename: meta.Filename,
		BucketId: bucketId,
		KeyId:    keyId,
	}
}

func NewSQLStore(genericDb gorm.Dialector) (*SQLStore, error) {
	db, err := gorm.Open(genericDb, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&SQLMetadata{})
	if err != nil {
		return nil, err
	}
	ret := new(SQLStore)
	ret.db = db
	return ret, nil
}

func (s *SQLStore) NewBucket(bucket string) error {
	s.m.Lock()
	defer s.m.Unlock()
	s.buckets = append(s.buckets, bucket)
	return nil
}

func (s *SQLStore) Write(bucketId, keyId string, metadata *Metadata) error {
	s.m.Lock()
	defer s.m.Unlock()
	var sqlMeta SQLMetadata
	_ = s.db.Where("bucket_id = ?", bucketId).Where("key_id = ?", keyId).Delete(&sqlMeta)
	sqlMetaRecord := SQLMetadataFromMetadata(bucketId, keyId, metadata)
	result := s.db.Create(sqlMetaRecord)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *SQLStore) Read(bucketId, keyId string) (*Metadata, error) {
	s.m.Lock()
	defer s.m.Unlock()
	sqlMeta := &SQLMetadata{}
	result := s.db.Where("bucket_id = ?", bucketId).Where("key_id = ?", keyId).First(sqlMeta)
	if result.Error != nil {
		return nil, result.Error
	}
	meta := MetadataFromSQLMetadata(sqlMeta)
	return meta, nil
}

func (s *SQLStore) Delete(bucket, key string) error {
	s.m.Lock()
	defer s.m.Unlock()
	var sqlMeta SQLMetadata
	result := s.db.Where("bucket_id = ?", bucket).Where("key_id = ?", key).Delete(&sqlMeta)
	return result.Error
}

func (s *SQLStore) DeleteBucket(bucket string) error {
	s.m.Lock()
	defer s.m.Unlock()
	var sqlMeta SQLMetadata
	result := s.db.Where("bucket_id = ?", bucket).Delete(&sqlMeta)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return result.Error
}

func (s *SQLStore) ListBuckets() ([]string, error) {
	s.m.Lock()
	defer s.m.Unlock()
	return s.buckets, nil
}
