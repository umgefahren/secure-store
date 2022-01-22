package access

import (
	"context"
	"crypto/sha512"
	"errors"
	"fmt"
	"time"
)

type AKey struct {
	expires         bool
	ttl             *time.Time
	limited         bool
	limit           uint64
	usedTimes       *uint64
	BucketId        string
	KeyId           string
	UrlKey          string
	NeedsKey        bool
	validKeys       *HashSet
	ResolveByUrlKey bool
}

type ExAccessKey struct {
	Ttl             *time.Time `json:"Ttl,omitempty"`
	Limit           *uint64    `json:"Limit,omitempty"`
	ValidKeys       [][]byte   `json:"ValidKeys,omitempty"`
	ResolveByUrlKey bool       `json:"ResolveByUrlKey,omitempty"`
	BucketId        string     `json:"BucketId"`
	KeyId           string     `json:"KeyId"`
	UrlKey          string     `json:"UrlKey"`
}

func FromExAccessKey(ex ExAccessKey) (*AKey, error) {
	validKeys := newHashSet()
	if len(ex.ValidKeys) == 0 {
		validKeys = nil
	} else {
		for _, e := range ex.ValidKeys {
			if len(e) != sha512.Size {
				return nil, errors.New("one of the valid keys doesn't has the wrong length")
			}
			arr := [sha512.Size]byte{}
			for i := range arr {
				arr[i] = e[i]
			}
			validKeys.Insert(arr)
		}
	}
	keyOpts, err := NewKeyOptions(ex.Ttl, ex.Limit, validKeys, ex.ResolveByUrlKey)
	if err != nil {
		return nil, err
	}
	ret := NewAccessKey(ex.BucketId, ex.KeyId, ex.UrlKey, *keyOpts)
	return ret, nil

}

type KeyOptions struct {
	expires         bool
	ttl             *time.Time
	limited         bool
	limit           uint64
	needsKey        bool
	validKeys       *HashSet
	resolveByUrlKey bool
}

func NewKeyOptions(ttl *time.Time, limit *uint64, validKeys *HashSet, resolveByUrlKey bool) (*KeyOptions, error) {
	if ttl != nil && time.Now().After(*ttl) {
		return nil, TTLAlreadyExpired(ttl)
	}
	ret := new(KeyOptions)
	ret.expires = ttl != nil
	ret.ttl = ttl
	ret.limited = limit != nil
	if limit == nil {
		ret.limit = 0
	} else {
		ret.limit = *limit
	}
	ret.needsKey = validKeys != nil
	ret.validKeys = validKeys
	ret.resolveByUrlKey = resolveByUrlKey

	return ret, nil
}

func NewAccessKey(bucketId, keyId, urlKey string, options KeyOptions) *AKey {
	zero := uint64(0)
	ret := new(AKey)
	ret.expires = options.expires
	ret.ttl = options.ttl
	ret.limited = options.limited
	ret.limit = options.limit
	ret.usedTimes = &zero
	ret.BucketId = bucketId
	ret.KeyId = keyId
	ret.UrlKey = urlKey
	ret.NeedsKey = options.needsKey
	ret.validKeys = options.validKeys
	ret.ResolveByUrlKey = options.resolveByUrlKey
	return ret
}

type AccessStore interface {
	AddKey(key *AKey) error
	Access(ctx context.Context, urlKey string) (chan<- bool, *AKey, error)
	DeleteKey(key *AKey) error
}

type byteArray [sha512.Size]byte

type hashSetMap map[byteArray]*interface{}

type HashSet struct {
	set hashSetMap
}

func (m *HashSet) Insert(obj byteArray) {
	m.set[obj] = nil
}

func (m *HashSet) Contains(obj byteArray) bool {
	_, ok := m.set[obj]
	return ok
}

func (m *HashSet) Delete(obj byteArray) {
	delete(m.set, obj)
}

func (m *HashSet) Size() int {
	return len(m.set)
}

func newHashSet() *HashSet {
	set := make(hashSetMap)
	ret := new(HashSet)
	ret.set = set
	return ret
}

func KeyAlreadyExists(urlKey string) error {
	return errors.New(fmt.Sprintf("key assigned to url key %v already exists", urlKey))
}

func KeyDoesntExist(urlKey string) error {
	return errors.New(fmt.Sprintf("no key is assigned to the url key %v", urlKey))
}

func TTLAlreadyExpired(ttl *time.Time) error {
	return errors.New(fmt.Sprintf("TTL %v already expired %v ago", ttl, time.Now().Sub(*ttl)))
}

func (a *AKey) ValidKey(key []byte) (bool, error) {
	if len(key) != sha512.Size {
		return false, errors.New("one of the valid keys doesn't has the wrong length")
	}
	arr := [sha512.Size]byte{}
	for i := range arr {
		arr[i] = key[i]
	}
	return a.validKeys.Contains(arr), nil
}
