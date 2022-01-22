package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"secure-store/access"
	"strings"
	"time"
)

type SecureClient struct {
	addr string
}

func NewClient(addr string) *SecureClient {
	client := new(SecureClient)
	client.addr = addr
	return client
}

func (s *SecureClient) CreateBucket(bucketId string) error {
	complete := fmt.Sprintf("%v/new-bucket?bucketId=%v", s.addr, bucketId)
	resp, err := http.Get(complete)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("server returned unexpected status code")
	}
	return nil
}

func (s *SecureClient) Ping() error {
	complete := fmt.Sprintf("%v/ping", s.addr)
	resp, err := http.Get(complete)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("returned unexpected response")
	}
	bodyString := new(strings.Builder)
	_, err = io.Copy(bodyString, resp.Body)
	if err != nil {
		return err
	}
	if bodyString.String() != "Pong" {
		return errors.New("server didn't responded with pong")
	}
	return nil
}

func (s *SecureClient) Upload(bucketId, keyId string, data io.Reader, filename string) error {
	complete := fmt.Sprintf("%v/upload?bucketId=%v&keyId=%v", s.addr, bucketId, keyId)
	req, err := http.NewRequest(http.MethodPost, complete, data)
	if err != nil {
		return err
	}
	req.Header.Add("filename", filename)
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	d, err := io.ReadAll(resp.Body)
	str := string(d)
	fmt.Println(str)
	return nil
}

func (s *SecureClient) Download(bucketId, keyId string) (io.Reader, int64, error) {
	complete := fmt.Sprintf("%v/download?bucketId=%v&keyId=%v", s.addr, bucketId, keyId)
	resp, err := http.Get(complete)
	if err != nil {
		return nil, 0, err
	}
	return resp.Body, resp.ContentLength, nil
}

func (s *SecureClient) Delete(bucketId, keyId string) error {
	complete := fmt.Sprintf("%v/delete?bucketId=%v&keyId=%v", s.addr, bucketId, keyId)
	req, err := http.NewRequest(http.MethodDelete, complete, nil)
	if err != nil {
		return err
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("unexpected server respond")
	}
	return nil
}

func (s *SecureClient) DeleteBucket(bucketId string) error {
	complete := fmt.Sprintf("%v/delete-bucket?bucketId=%v", s.addr, bucketId)
	req, err := http.NewRequest(http.MethodDelete, complete, nil)
	if err != nil {
		return err
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("unexpected server respond")
	}
	return nil
}

func (s *SecureClient) AddKey(ttl *time.Time, limit *uint64, validKeys [][]byte, resolveByUrlKey bool, bucketId, keyId, urlKey string) error {
	exKey := &access.ExAccessKey{
		Ttl:             ttl,
		Limit:           limit,
		ValidKeys:       validKeys,
		ResolveByUrlKey: resolveByUrlKey,
		BucketId:        bucketId,
		KeyId:           keyId,
		UrlKey:          urlKey,
	}
	complete := fmt.Sprintf("%v/api/add", s.addr)
	jsonBytes, err := json.Marshal(exKey)
	if err != nil {
		return err
	}
	requestBody := bytes.NewBuffer(jsonBytes)
	req, err := http.NewRequest(http.MethodPost, complete, requestBody)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("unexpected server respond")
	}
	return nil
}

func (s *SecureClient) DownloadFromKey(urlKey string, unlockKey []byte, inUrlKey bool) (io.Reader, int64, error) {
	complete := ""
	base64UnlockKey := base64.RawURLEncoding.EncodeToString(unlockKey)
	if inUrlKey {
		complete = fmt.Sprintf("%v/api/download?urlKey=%v&unlockKey=%v", s.addr, urlKey, base64UnlockKey)
	} else {
		complete = fmt.Sprintf("%v/api/download?urlKey=%v", s.addr, urlKey)
	}
	req, err := http.NewRequest(http.MethodGet, complete, nil)
	if err != nil {
		return nil, 0, err
	}
	if !inUrlKey {
		req.Header.Add("unlockKey", base64UnlockKey)
	}
	client := http.DefaultClient
	log.Println("Initialized client")
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, errors.New("unexpected server response")
	}
	return resp.Body, resp.ContentLength, nil
}
