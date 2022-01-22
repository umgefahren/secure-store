package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"secure-store/access"
	"secure-store/metadata"
	"secure-store/security"
	"strings"
	"time"
)

const UnlockKeyQuery = "unlockKey"

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
}

func NewRouter(s *CompoundStore, a access.AccessStore) *gin.Engine {
	matcher := NewMatcher()

	router := gin.Default()

	isInDebugMode := os.Getenv(gin.EnvGinMode) == "" || strings.ToLower(os.Getenv(gin.EnvGinMode)) == gin.DebugMode

	if isInDebugMode {
		log.Println("Mode: Debug")
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
			return
		}
		c.String(http.StatusOK, "Created bucket with id: %v", bucketId)
	})

	router.POST("/upload", func(c *gin.Context) {
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
		err := s.Write(bucketId, keyId, meta, key, bufferedReader)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.String(http.StatusOK, "Successfully uploaded to bucket-id: %v with key-id: %v", bucketId, keyId)
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
			return
		}
		c.String(http.StatusOK, "Deleted with buket-id: %v and key-id: %v", bucketId, keyId)
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
			return
		}
		c.String(http.StatusOK, "Deleted bucket-id: %v", bucketId)
	})

	router.POST("/api/add", func(c *gin.Context) {
		exKey := &access.ExAccessKey{}
		contentType := c.Request.Header.Get("Content-Type")
		if contentType != "application/json" {
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("wrong content type given"))
			return
		}
		err := c.ShouldBindJSON(exKey)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		key, err := access.FromExAccessKey(*exKey)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		err = a.AddKey(key)
		if err != nil {
			_ = c.AbortWithError(http.StatusBadRequest, err)
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
				_ = c.AbortWithError(http.StatusInternalServerError, err)
			} else {
				_ = c.AbortWithError(http.StatusForbidden, AccessForbiddenError())
			}
			return
		}
		if err != nil {
			if isInDebugMode {
				_ = c.AbortWithError(http.StatusInternalServerError, access.KeyDoesntExist(urlKey))
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
			} else {
				_ = c.AbortWithError(http.StatusForbidden, AccessForbiddenError())
			}
			return
		}
		bucketId, keyId := key.BucketId, key.KeyId
		Download(c, s, bucketId, keyId)
		backMessage = true
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
	})

	return router
}
