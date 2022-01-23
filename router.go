package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	rainbow "github.com/guineveresaenger/golang-rainbow"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"secure-store/access"
	"secure-store/metadata"
	"secure-store/security"
	"secure-store/users"
	"strings"
	"time"
)

const UnlockKeyQuery = "unlockKey"
const ApiKeyQuery = "apiKey"

func AccessForbiddenError() error {
	return errors.New("access forbidden")
}

func Download(ctx *gin.Context, s *CompoundStore, bucketId, keyId string) {
	meta, bufReader, err := s.Read(bucketId, keyId)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	contentLength := meta.Length
	contentType := "application/octet-stream"
	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%v"`, meta.Filename),
	}
	ctx.DataFromReader(http.StatusOK, contentLength, contentType, bufReader, extraHeaders)
	logrus.WithFields(logrus.Fields{
		"Bucket Id":      bucketId,
		"Key Id":         keyId,
		"Content Length": contentLength,
		"Filename":       meta.Filename,
	}).Infoln("Successfully downloaded.")
}

func NewRouter(s *CompoundStore, a access.AccessStore, u users.UserStorage) *gin.Engine {
	matcher := NewMatcher()

	router := gin.New()
	// router.Use(ginlogrus.Logger(logrus.New()), gin.Recovery())
	router.Use(gin.Recovery(), gin.Logger())
	isInDebugMode := os.Getenv(gin.EnvGinMode) == "" || strings.ToLower(os.Getenv(gin.EnvGinMode)) == gin.DebugMode

	if isInDebugMode {
		logrus.Infoln("Using Debug mode")
	}

	router.LoadHTMLGlob("templates/*")

	router.GET("/new-bucket", func(c *gin.Context) {
		bucketId := c.Query("bucketId")
		matchRes := matcher.MatchString(bucketId)
		if !matchRes {
			_ = c.AbortWithError(http.StatusBadRequest, BucketIdMatchingError())
			return
		}
		err := s.NewBucket(bucketId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			logrus.WithError(err).Errorf("Error while creating a new bucket.")
			return
		}
		c.String(http.StatusOK, "Created bucket with id: %v", bucketId)
		logrus.WithFields(logrus.Fields{
			"Bucket Id": bucketId,
		}).Infoln("Successfully created a new bucket.")
	})

	router.POST("/upload", func(c *gin.Context) {
		apiKey := c.Query(ApiKeyQuery)
		apiKeyBytes, err := base64.RawURLEncoding.DecodeString(apiKey)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		user, err := u.ResolveByApiKey(apiKeyBytes)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		verifiedKey := user.VerifyApiKey(apiKeyBytes)
		if !verifiedKey {
			_ = c.AbortWithError(http.StatusForbidden, AccessForbiddenError())
			return
		}
		if !user.Role.CanUploadData {
			_ = c.AbortWithError(http.StatusForbidden, AccessForbiddenError())
			return
		}
		bucketId := c.Query("bucketId")
		matchRes := matcher.MatchString(bucketId)
		if !matchRes {
			_ = c.AbortWithError(http.StatusBadRequest, BucketIdMatchingError())
			return
		}
		keyId := c.Query("keyId")
		matchRes = matcher.MatchString(keyId)
		if !matchRes {
			_ = c.AbortWithError(http.StatusBadRequest, KeyIdMatchingError())
			return
		}
		r := c.Request.Body
		filename := c.Request.Header.Get("filename")
		contentLength := c.Request.ContentLength
		bufferedReader := bufio.NewReader(r)
		meta := metadata.NewMetadata(contentLength, filename)
		key := security.NewEncryptionKey()
		err = s.Write(bucketId, keyId, meta, key, bufferedReader)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			logrus.WithError(err).Errorf("Error while writing data into storage.")
			return
		}
		c.String(http.StatusOK, "Successfully uploaded to bucket-id: %v with key-id: %v", bucketId, keyId)
		logrus.WithFields(logrus.Fields{
			"Bucket Id":      bucketId,
			"Key Id":         keyId,
			"Filename":       filename,
			"Content Length": contentLength,
		}).Infoln("Successfully uploaded.")
	})

	router.GET("/download", func(c *gin.Context) {
		bucketId := c.Query("bucketId")
		matchRes := matcher.MatchString(bucketId)
		if !matchRes {
			_ = c.AbortWithError(http.StatusBadRequest, BucketIdMatchingError())
			return
		}
		keyId := c.Query("keyId")
		matchRes = matcher.MatchString(keyId)
		if !matchRes {
			_ = c.AbortWithError(http.StatusBadRequest, KeyIdMatchingError())
			return
		}
		Download(c, s, bucketId, keyId)
	})

	router.DELETE("/delete", func(c *gin.Context) {
		bucketId := c.Query("bucketId")
		matchRes := matcher.MatchString(bucketId)
		if !matchRes {
			_ = c.AbortWithError(http.StatusBadRequest, BucketIdMatchingError())
			return
		}
		keyId := c.Query("keyId")
		matchRes = matcher.MatchString(keyId)
		if !matchRes {
			_ = c.AbortWithError(http.StatusBadRequest, KeyIdMatchingError())
			return
		}
		err := s.Delete(bucketId, keyId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			logrus.WithError(err).WithFields(logrus.Fields{
				"Bucket Id": bucketId,
				"Key Id":    keyId,
			}).Errorf("Error encountered during delete")
			return
		}
		c.String(http.StatusOK, "Deleted with buket-id: %v and key-id: %v", bucketId, keyId)
		logrus.WithFields(logrus.Fields{
			"Bucket Id": bucketId,
			"Key Id":    keyId,
		}).Debugf("Successfully deleted object.")
	})

	router.DELETE("/delete-bucket", func(c *gin.Context) {
		bucketId := c.Query("bucketId")
		matchRes := matcher.MatchString(bucketId)
		if !matchRes {
			_ = c.AbortWithError(http.StatusBadRequest, BucketIdMatchingError())
			return
		}
		err := s.DeleteBucket(bucketId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			logrus.WithError(err).Errorf("During deletion of bucket.")
			return
		}
		c.String(http.StatusOK, "Deleted bucket-id: %v", bucketId)
		logrus.WithFields(logrus.Fields{
			"Bucket Id": bucketId,
		}).Debugf("Successfully deleted bucket.")
	})

	router.POST("/api/add", func(c *gin.Context) {
		apiKey := c.Query(ApiKeyQuery)
		apiKeyBytes, err := base64.RawURLEncoding.DecodeString(apiKey)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		user, err := u.ResolveByApiKey(apiKeyBytes)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		verifiedKey := user.VerifyApiKey(apiKeyBytes)
		if !verifiedKey {
			_ = c.AbortWithError(http.StatusForbidden, AccessForbiddenError())
			return
		}
		if !user.Role.CanAddKeys {
			_ = c.AbortWithError(http.StatusForbidden, AccessForbiddenError())
			return
		}
		exKey := &access.ExAccessKey{}
		contentType := c.Request.Header.Get("Content-Type")
		if contentType != "application/json" {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("wrong content type given"))
			logrus.WithError(errors.New("wrong content type given")).Errorf("During request the wrong content type was given.")
			return
		}
		err = c.ShouldBindJSON(exKey)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			logrus.WithError(err).Errorf("Unsuccessfully binded into JSON.")
			return
		}
		key, err := access.FromExAccessKey(*exKey)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			logrus.WithError(err).Errorf("Unsuccessfully accessed.")
			return
		}
		err = a.AddKey(key)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			logrus.WithError(err).Errorf("Unsuccessfully accessed key.")
			return
		}
		c.String(http.StatusOK, "access key was set")
	})

	router.GET("/api/download", func(c *gin.Context) {
		ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(100*time.Second))
		defer cancel()
		urlKey := c.Query("urlKey")
		matchRes := matcher.MatchString(urlKey)
		if !matchRes {
			if isInDebugMode {
				_ = c.AbortWithError(http.StatusBadRequest, UrlKeyMatchingError())
			} else {
				_ = c.AbortWithError(http.StatusForbidden, AccessForbiddenError())
			}
			return
		}
		back, key, err := a.Access(ctx, urlKey)
		backMessage := false
		defer func() {
			go func() {
				back <- backMessage
			}()
		}()
		if err != nil {
			if isInDebugMode {
				_ = c.AbortWithError(http.StatusInternalServerError, access.KeyDoesntExist(urlKey))
				logrus.WithError(err).Errorf("Key doesn't of key exists.")
			} else {
				_ = c.AbortWithError(http.StatusForbidden, AccessForbiddenError())
			}
			return
		}
		if !key.NeedsKey {
			Download(c, s, key.BucketId, key.KeyId)
			backMessage = true
			return
		}
		unlockKey := make([]byte, 0)
		if key.ResolveByUrlKey {
			queryKey := c.Query(UnlockKeyQuery)
			unlockKey, err = base64.RawURLEncoding.DecodeString(queryKey)
			if err != nil {
				if isInDebugMode {
					_ = c.AbortWithError(http.StatusBadRequest, errors.New("key in query could not be base64 decoded"))
					logrus.WithError(err).Errorf("Decoding of key failed.")
				} else {
					_ = c.AbortWithError(http.StatusForbidden, AccessForbiddenError())
				}
				return
			}
		} else {
			queryKey := c.GetHeader(UnlockKeyQuery)
			unlockKey, err = base64.RawURLEncoding.DecodeString(queryKey)

			if err != nil {
				if isInDebugMode {
					_ = c.AbortWithError(http.StatusBadRequest, errors.New("key in header could not be base64 decoded"))
					logrus.WithError(err).Errorf("Decoding of key failed.")
				} else {
					_ = c.AbortWithError(http.StatusForbidden, AccessForbiddenError())
				}
				return
			}
		}
		keyValid, err := key.ValidKey(unlockKey)
		if !keyValid {
			if isInDebugMode {
				_ = c.AbortWithError(http.StatusForbidden, errors.New("key is invalid"))
				logrus.WithError(err).Errorf("Key is invalid.")
			} else {
				_ = c.AbortWithError(http.StatusForbidden, AccessForbiddenError())
			}
			return
		}
		bucketId, keyId := key.BucketId, key.KeyId
		Download(c, s, bucketId, keyId)
		backMessage = true
	})

	router.POST("/api/user/create", func(c *gin.Context) {
		apiKey := c.Query(ApiKeyQuery)
		apiKeyBytes, err := base64.RawURLEncoding.DecodeString(apiKey)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		rootUser, err := u.ResolveByApiKey(apiKeyBytes)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if !rootUser.Role.CanCreateUsers || !rootUser.Role.RootUser {
			_ = c.AbortWithError(http.StatusForbidden, AccessForbiddenError())
		}
		userJson := &users.UserJson{}
		contentType := c.Request.Header.Get("Content-Type")
		if contentType != "application/json" {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("wrong content type given"))
			logrus.WithError(errors.New("wrong content type given")).Errorf("During request the wrong content type was given.")
			return
		}
		err = c.ShouldBindJSON(userJson)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		err = userJson.IsValid()
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		user, err := users.UserFromUserJson(userJson)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		err = u.Create(user)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		exUser := users.UserSafeJsonFromUser(user)
		c.SecureJSON(http.StatusOK, exUser)
	})

	router.GET("/teapot", func(c *gin.Context) {
		ip, _ := c.RemoteIP()
		t := time.Now()
		routes := router.Routes()
		routesString := make([]string, 0)
		for _, route := range routes {
			routesString = append(routesString, route.Path)
		}
		timeString := t.Format(time.RFC850)
		c.HTML(http.StatusTeapot, "teapot.tmpl", gin.H{
			"title":     "I'm a teapot",
			"ipAddress": ip.String(),
			"time":      timeString,
			"Routes":    routesString,
		})
		rainbow.Rainbow("YoU hAvE fOuNd ThE tEaPoT!!!", 0)
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{})
	})

	router.StaticFile("/favicon.ico", "./static/favicon.ico")

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "Pong")
	})

	return router
}
