package main

import (
	"context"
	"crypto/sha512"
	"fmt"
	"github.com/gin-gonic/autotls"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"secure-store/access"
	"secure-store/metadata"
	"secure-store/security"
	"secure-store/storage"
	"secure-store/users"
	"strconv"
	"strings"
)

const PortEnv = "PORT"
const DataDirEnv = "DATA_DIR"
const StorageEnv = "STORAGE_KIND"
const AccessEnv = "ACCESS_KIND"

const StorageEnvFs = "FS_STORAGE"
const StorageEnvMem = "MEM_STORAGE"

const AccessEnvMem = "MEM_ACCESS"
const AccessEnvRedis = "REDIS_ACCESS"

const RedisEnvHost = "REDIS_HOST"
const RedisEnvPort = "REDIS_PORT"
const RedisEnvPassword = "REDIS_PASSWORD"
const RedisEnvDb = "REDIS_DB"

const RootUserId = "83672c3d-bb08-4d65-9d71-1191dc11cb80"
const RootName = "Root"
const RootUsername = "root"

var RootPassword = []byte("password")

var HashedRootPassword = sha512.New().Sum(RootPassword)

var RootRule = users.Role{
	RootUser:       true,
	CanCreateUsers: true,
	CanAddKeys:     true,
	CanUploadData:  true,
	CanDeleteKeys:  true,
}

var RootApiKey = []byte("root-api-key")

var RootUserJson = users.UserJson{
	Id:           RootUserId,
	Name:         RootName,
	Role:         RootRule,
	Username:     RootUsername,
	PasswordHash: HashedRootPassword,
	ApiKey:       RootApiKey,
}

var RootUser, _ = users.UserFromUserJson(&RootUserJson)

func main() {
	fmt.Println("Welcome to Secure-Store v0.0.1 👋")
	fmt.Println("I will keep your files secure and accessible 🔒")
	fmt.Println(" ❌  Don't use this software in production!! ❌  ")
	fmt.Println("@umgefahren")

	port := os.Getenv(PortEnv)
	if port == "" {
		port = "8080"
	}

	logrus.WithField("port", port).Info("Starting server")

	var s storage.Storage
	storageEnv := os.Getenv(StorageEnv)
	logrus.WithFields(logrus.Fields{
		"Storage Env": storageEnv,
	}).Infoln("Storage Env variable")
	switch storageEnv {
	case StorageEnvFs:
		dataDir := os.Getenv(DataDirEnv)
		if dataDir == "" {
			dataDir = os.TempDir() + "data/" + uuid.NewString()
		}

		logrus.WithField("Data Directory", dataDir).Info("Starting server with data directory paramter")
		store, err := storage.NewFsStorage(dataDir)
		if err != nil {
			log.Fatal(err)
		}
		logrus.Infoln("Using FS Storage")
		s = store
	case StorageEnvMem:
		store := storage.NewMemoryStorage()
		logrus.Infoln("Using Memory Storage")
		s = store
	default:
		err := os.Setenv(StorageEnv, StorageEnvMem)
		if err != nil {
			logrus.WithError(err).Errorf("While setting environment variable")
		}
		logrus.WithField("Storage Env", storageEnv).Fatal("storage env variable is invalid, setting storage to memory")
	}

	m := metadata.NewMemoryStore()
	sec := security.NewMemorySecurityStore()
	compound := CompoundStore{
		metadata: m,
		security: sec,
		storage:  s,
	}

	var a access.AccessStore
	accessEnv := os.Getenv(AccessEnv)
	switch accessEnv {
	case AccessEnvMem:
		a = access.NewMemoryStore()
		logrus.Infoln("Using in memory access storage")
	case AccessEnvRedis:
		redisHost := os.Getenv(RedisEnvHost)
		if redisHost == "" {
			redisHost = "0.0.0.0"
			logrus.WithField("Redis Host", redisHost).Infoln("Falling back on default redis host")
		}
		redisPort := os.Getenv(RedisEnvPort)
		if redisPort == "" {
			redisPort = "6379"
			logrus.WithField("Redis Port", redisPort).Infoln("Falling back on default redis port")
		}
		redisPassword := os.Getenv(RedisEnvPassword)
		if redisPassword == "" {
			redisPassword = ""
			logrus.WithField("Redis Password", redisPassword).Infoln("Falling back on default redis password")
		}
		redisDb := os.Getenv(RedisEnvDb)
		if redisDb == "" {
			redisDb = "0"
			logrus.WithField("Redis Db", redisDb).Infoln("Falling back on default redis db")
		}
		redisDbInt, err := strconv.ParseInt(redisDb, 10, 32)
		if err != nil {
			logrus.WithError(err).Fatal("Couldn't parse redis db environment variable")
		}
		logrus.WithFields(logrus.Fields{
			"Redis Host": redisHost,
			"Redis Port": redisPort,
			"Redis Db":   redisDb,
		}).Infoln("Starting redis client.")
		redisClient := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%v:%v", redisHost, redisPort),
			Password: redisPassword,
			DB:       int(redisDbInt),
		})
		accessStore, err := access.NewRedisStore(context.TODO(), redisClient)
		if err != nil {
			log.Fatal(err)
		}
		a = accessStore
	default:
		err := os.Setenv(AccessEnv, AccessEnvMem)
		if err != nil {
			logrus.WithError(err).Errorf("While setting environment variable")
		}
		logrus.WithField("Access Env", accessEnv).Fatal("access env variable is invalid, setting storage to memory")
	}

	u := users.NewMemoryStore()
	err := u.Create(RootUser)
	if err != nil {
		logrus.WithError(err).Infoln("Failed during adding root user")
	} else {
		logrus.Infoln("Added root user")
	}
	r := NewRouter(&compound, a, u)

	domainsString := os.Getenv("DOMAINS")
	if domainsString == "" {
		logrus.Fatal(r.Run(fmt.Sprintf("0.0.0.0:%v", port)))
	} else {
		domains := strings.Split(domainsString, ",")
		domainsFiltered := make([]string, 0)
		for _, d := range domains {
			if d != "" {
				domainsFiltered = append(domainsFiltered, d)
			}
		}
		logrus.Fatal(autotls.Run(r, domainsFiltered...))
	}
}
