package main

import (
	"bytes"
	"fmt"
	"io"
	"secure-store/client"
)

func main() {
	c := client.NewClient("http://localhost:8080")
	err := c.CreateBucket("lol")
	if err != nil {
		panic(err)
	}
	data := []byte("Hello World")
	r := bytes.NewReader(data)
	err = c.Upload("lol", "yea", r, "lol.txt")
	if err != nil {
		panic(err)
	}
	respR, err := c.Download("lol", "yea")
	if err != nil {
		panic(err)
	}
	respData, err := io.ReadAll(respR)
	if err != nil {
		panic(err)
	}
	respString := string(respData)
	fmt.Println(respString)
	err = c.Delete("lol", "yea")
	if err != nil {
		panic(err)
	}
	err = c.DeleteBucket("lol")
	if err != nil {
		panic(err)
	}
}
