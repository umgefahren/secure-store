package metadata

type MetadataStore interface {
	NewBucket(bucket string) error
	Write(bucketId, keyId string, metadata *Metadata) error
	Read(bucketId, keyId string) (*Metadata, error)
	Delete(bucket, key string) error
	DeleteBucket(bucket string) error
	ListBuckets() ([]string, error)
}

type Metadata struct {
	Length   int64
	Filename string
}

func NewMetadata(length int64, filename string) *Metadata {
	ret := new(Metadata)
	ret.Length = length
	ret.Filename = filename
	return ret
}
