package access

import (
	"context"
	"crypto/sha512"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func ProtoKeyFromAKey(key *AKey) *ProtoAKey {
	ret := new(ProtoAKey)
	ret.Expires = key.expires
	if key.ttl == nil {
		ret.Ttl = nil
	} else {
		ret.Ttl = timestamppb.New(*key.ttl)
	}
	ret.Limited = key.limited
	ret.Limit = key.limit
	if key.usedTimes == nil {
		ret.UsedTimes = 0
	} else {
		ret.UsedTimes = *key.usedTimes
	}
	ret.BucketId = key.BucketId
	ret.KeyId = key.KeyId
	ret.UrlKey = key.UrlKey
	ret.NeedsKey = key.NeedsKey

	keys := make([][]byte, 0)
	if ret.NeedsKey {
		for array := range key.validKeys.set {
			keys = append(keys, array[:])
		}
	}

	ret.ValidKeys = keys
	ret.ResolveByUrlKey = key.ResolveByUrlKey
	return ret
}

func AKeyFromProtoKey(proto *ProtoAKey) *AKey {
	ret := new(AKey)
	ret.expires = proto.Expires
	if proto.Ttl == nil {
		ret.ttl = nil
	} else {
		ttl := proto.Ttl.AsTime()
		ret.ttl = &ttl
	}
	ret.limited = proto.Limited
	ret.limit = proto.Limit
	ret.usedTimes = &proto.UsedTimes
	ret.BucketId = proto.BucketId
	ret.KeyId = proto.KeyId
	ret.UrlKey = proto.UrlKey
	ret.NeedsKey = proto.NeedsKey

	keys := new(HashSet)
	keys.set = make(hashSetMap)
	for _, key := range proto.ValidKeys {
		validKey := [sha512.Size]byte{}
		for i := 0; i < sha512.Size; i++ {
			validKey[i] = key[i]
		}
		keys.Insert(validKey)
	}
	ret.validKeys = keys
	ret.ResolveByUrlKey = proto.ResolveByUrlKey
	return ret
}

type RedisStore struct {
	client     *redis.Client
	ctx        context.Context
	killerChan chan<- *AKey
}

func NewRedisStore(ctx context.Context, client *redis.Client) (*RedisStore, error) {
	statusCmd := client.Ping(ctx)
	err := statusCmd.Err()
	if err != nil {
		return nil, err
	}
	ret := new(RedisStore)
	ret.client = client
	ret.ctx = ctx
	killerChan := make(chan *AKey)
	ret.killerChan = killerChan
	go func() {
		for {
			k := <-killerChan
			if k == nil {
				return
			}
			_ = ret.DeleteKey(k)
		}
	}()
	return ret, nil
}

func (r *RedisStore) AddKey(key *AKey) error {
	protoKey := ProtoKeyFromAKey(key)
	data, err := proto.Marshal(protoKey)
	if err != nil {
		return err
	}
	statusCmd := &redis.StatusCmd{}
	if key.expires {
		statusCmd = r.client.Set(r.ctx, key.UrlKey, data, time.Now().Sub(*key.ttl))
	} else {
		statusCmd = r.client.Set(r.ctx, key.UrlKey, data, time.Second*0)
	}
	if statusCmd.Err() != nil {
		return statusCmd.Err()
	}
	return nil
}

func (r *RedisStore) Access(ctx context.Context, urlKey string) (chan<- bool, *AKey, error) {
	backChan := make(chan bool)
	statusCmd := r.client.Get(ctx, urlKey)
	data, err := statusCmd.Bytes()
	if err != nil {
		return nil, nil, err
	}
	protoKey := &ProtoAKey{}
	err = proto.Unmarshal(data, protoKey)
	if err != nil {
		return nil, nil, err
	}
	aKey := AKeyFromProtoKey(protoKey)
	go func() {
		select {
		case <-ctx.Done():
			return
		case resp := <-backChan:
			statusCmd = r.client.Get(ctx, urlKey)
			if statusCmd.Err() != nil {
				logrus.WithError(err).WithField("Url Key", urlKey).Errorf("error retrieving data from redis")
				return
			}
			protoData, err := statusCmd.Bytes()
			if err != nil {
				logrus.WithError(err).Errorf("error while retrieving bytes from Redis back")
				return
			}
			protoAKey := &ProtoAKey{}
			err = proto.Unmarshal(protoData, protoAKey)
			if err != nil {
				logrus.WithError(err).Errorf("unmarshiling in coroutine failed")
				return
			}
			if resp {
				protoAKey.UsedTimes += 1
				protoData, err = proto.Marshal(protoAKey)
				if err != nil {
					logrus.WithError(err).WithField("Url Key", urlKey).Errorf("Error while marshalling")
					return
				}
				r.client.Set(ctx, urlKey, protoData, time.Now().Sub(protoAKey.Ttl.AsTime()))
			}
			if protoAKey.UsedTimes >= protoAKey.Limit && protoAKey.Limited {
				r.killerChan <- AKeyFromProtoKey(protoAKey)
			}
			return
		}
	}()
	return backChan, aKey, nil
}

func (r *RedisStore) DeleteKey(key *AKey) error {
	statusCmd := r.client.Del(r.ctx, key.UrlKey)
	if statusCmd.Err() != nil {
		return statusCmd.Err()
	}
	return nil
}
