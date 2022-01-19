package main

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"secure-store/metadata"
	"secure-store/storage"
)

func NewRouter(s storage.Storage, m metadata.MetadataStore) *gin.Engine {
	router := gin.Default()

	router.GET("/new-bucket", func(c *gin.Context) {
		bucketId := c.Query("bucketId")
		err := s.NewBucket(bucketId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}
		err = m.NewBucket(bucketId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
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
		err := s.Write(bucketId, keyId, bufferedReader)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}
		meta := metadata.NewMetadata(contentLength, filename)
		err = m.Write(bucketId, keyId, meta)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}
		c.String(http.StatusOK, "Successfully uploaded to bucket-id: %v with key-id: %v", bucketId, keyId)
	})

	router.GET("/download", func(c *gin.Context) {
		bucketId := c.Query("bucketId")
		keyId := c.Query("keyId")
		bufReader, err := s.Read(bucketId, keyId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}
		meta, err := m.Read(bucketId, keyId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
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
		}
		err = m.Delete(bucketId, keyId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}
		c.String(http.StatusOK, "Deleted with buket-id: %v and key-id: %v", bucketId, keyId)
	})

	router.DELETE("/delete-bucket", func(c *gin.Context) {
		bucketId := c.Query("bucketId")
		err := s.DeleteBucket(bucketId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}
		err = m.DeleteBucket(bucketId)
		if err != nil {
			_ = c.AbortWithError(http.StatusInternalServerError, err)
		}
		c.String(http.StatusOK, "Deleted bucket-id: %v", bucketId)
	})

	/*
		router.StaticFile("/", "./demo-app/public/index.html")
		router.StaticFile("/global.css", "./demo-app/public/global.css")
		router.StaticFile("/favicon.png", "./demo-app/public/favicon.png")
		router.StaticFile("/build/bundle.css", "./demo-app/public/build/bundle.css")
		router.StaticFile("/build/bundle.js", "./demo-app/public/build/bundle.js")
		router.StaticFile("/build/bundle.js.map", "./demo-app/public/build/bundle.js.map")
	*/

	return router
}
