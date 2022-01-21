package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"secure-store/access"
	"secure-store/metadata"
	"secure-store/security"
	"time"
)

func NewRouter(s *CompoundStore, a access.AccessStore) *gin.Engine {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")

	router.GET("/new-bucket", func(c *gin.Context) {
		bucketId := c.Query("bucketId")
		err := s.NewBucket(bucketId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.String(http.StatusOK, "Created bucket with id: %v", bucketId)
	})

	router.POST("/upload", func(c *gin.Context) {
		bucketId := c.Query("bucketId")
		keyId := c.Query("keyId")
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
		keyId := c.Query("keyId")
		meta, bufReader, err := s.Read(bucketId, keyId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		contentLength := meta.Length
		contentType := "application/octet-stream"
		extraHeaders := map[string]string{
			"Content-Disposition": fmt.Sprintf(`attachment; filename="%v"`, meta.Filename),
		}
		c.DataFromReader(http.StatusOK, contentLength, contentType, bufReader, extraHeaders)
	})

	router.DELETE("/delete", func(c *gin.Context) {
		bucketId := c.Query("bucketId")
		keyId := c.Query("keyId")
		err := s.Delete(bucketId, keyId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.String(http.StatusOK, "Deleted with buket-id: %v and key-id: %v", bucketId, keyId)
	})

	router.DELETE("/delete-bucket", func(c *gin.Context) {
		bucketId := c.Query("bucketId")
		err := s.DeleteBucket(bucketId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.String(http.StatusOK, "Deleted bucket-id: %v", bucketId)
	})

	router.PUT("/api/add", func(c *gin.Context) {
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
	})

	router.GET("/api/access", func(c *gin.Context) {
		ctx, _ := context.WithDeadline(context.TODO(), time.Now().Add(100*time.Second))
		urlKey := c.Query("urlKey")
		_, _, _ = a.Access(ctx, urlKey)
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
