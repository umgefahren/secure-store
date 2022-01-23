package access

import (
	ctx "context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"testing"
)

const RedisTestEnvHost = "REDIS_TEST_HOST"
const RedisTestEnvPort = "REDIS_TEST_PORT"
const RedisTestEnvSkip = "REDIS_TEST_SKIP"

var redisHost = os.Getenv(RedisTestEnvHost)
var redisPort = os.Getenv(RedisTestEnvPort)

func TestAddAndDeleteRedis(t *testing.T) {
	_, ok := os.LookupEnv(RedisTestEnvSkip)
	if !ok {
		t.SkipNow()
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", redisHost, redisPort),
		Password: "",
		DB:       1,
	})
	db, err := NewRedisStore(ctx.TODO(), redisClient)
	if err != nil {
		t.Fatal(err)
	}
	AddAndDeleteTest(t, db)
}

func TestAddAndAccessRedis(t *testing.T) {
	_, ok := os.LookupEnv(RedisTestEnvSkip)
	if !ok {
		t.SkipNow()
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", redisHost, redisPort),
		Password: "",
		DB:       2,
	})
	db, err := NewRedisStore(ctx.TODO(), redisClient)
	if err != nil {
		t.Fatal(err)
	}
	AddAndAccessTest(t, db)
}
