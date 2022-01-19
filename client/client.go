package client

import (
	"errors"
	"fmt"
	"io"
	"net/http"
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
