package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/google/uuid"
	"github.com/manifoldco/promptui"
	"golang.org/x/crypto/argon2"
	"io"
	"log"
	"os"
	"secure-store/client"
	"secure-store/users"
	"strings"
	"time"
)

var RootApiKey = []byte("root-api-key")

func main() {
	c := client.NewClient("http://localhost:8080")
	items := []string{"Create Bucket", "Read", "Write", "Delete", "DeleteBucket", "Add Key", "Download From Key", "Add User", "Exit"}
	for {
		prompt := promptui.Select{
			Label:             "Select operation",
			StartInSearchMode: true,
			Searcher: func(input string, idx int) bool {
				val := items[idx]

				return strings.Contains(strings.ToLower(val), strings.ToLower(input))
			},
			Items: items,
		}
		op, opString, err := prompt.Run()
		if err != nil {
			log.Fatal(err)
		}
		if opString == "Exit" {
			break
		}
		bucket := ""
		if op != 6 && op != 7 {
			bucket = AskForBucketId()
		}
		switch op {
		case 0:
			err = c.CreateBucket(bucket)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Created bucket with id %v", bucket)
		case 1:
			key := AskForKeyId()
			data, length, err := c.Download(bucket, key)
			if err != nil {
				log.Println(err)
				continue
			}
			shouldContinue := DownloadInteraction(data, length)
			if shouldContinue {
				continue
			}
		case 2:
			targetPrompt := promptui.Select{
				Label: "From what should be read",
				Items: []string{"File", "Stdin"},
			}
			_, target, err := targetPrompt.Run()
			if err != nil {
				log.Fatal(err)
			}
			switch target {
			case "File":
				filePrompt := promptui.Prompt{
					Label: "From what file should be read?",
				}
				path, err := filePrompt.Run()
				if err != nil {
					log.Fatal(err)
				}
				file, err := os.Open(path)
				if err != nil {
					log.Println(err)
					continue
				}
				key := AskForKeyId()
				name := file.Name()
				stat, err := file.Stat()
				if err != nil {
					log.Fatal(err)
				}
				length := stat.Size()
				bar := pb.Full.Start64(length)
				barReader := bar.NewProxyReader(file)
				apiKey := []byte("api-key-key")
				apiKeyHash := sha512.Sum512(apiKey)
				err = c.Upload(bucket, key, barReader, name, apiKeyHash[:])
				if err != nil {
					log.Println(err)
					continue
				}
			case "Stdin":
				contentPrompt := promptui.Prompt{
					Label: "Content",
				}
				content, err := contentPrompt.Run()
				if err != nil {
					log.Fatal(err)
				}
				filenamePrompt := promptui.Prompt{
					Label: "Under which filename",
				}
				filename, err := filenamePrompt.Run()
				if err != nil {
					log.Fatal(err)
				}
				key := AskForKeyId()
				contentReader := strings.NewReader(content)
				apiKey := []byte("api-key-key")
				apiKeyHash := sha512.Sum512(apiKey)
				err = c.Upload(bucket, key, contentReader, filename, apiKeyHash[:])
				if err != nil {
					log.Println(err)
					continue
				}
			}
		case 3:
			key := AskForKeyId()
			err = c.Delete(bucket, key)
			if err != nil {
				log.Println(err)
				continue
			}
		case 4:
			err = c.DeleteBucket(bucket)
			if err != nil {
				log.Println(err)
				continue
			}
		case 5:
			key := AskForKeyId()
			var ttl *time.Time = nil
			var limit *uint64 = nil
			validKeys := make([][]byte, 0)
			urlKey := AskForUrlKey()
			if AskIfPassword() {
				password, err := GetPassword()
				if err == nil {
					validKeys = append(validKeys, password)
				}
			}
			apiKey := []byte("api-key-key")
			apiKeyHash := sha512.Sum512(apiKey)
			err = c.AddKey(ttl, limit, validKeys, true, bucket, key, urlKey, apiKeyHash[:])
			if err != nil {
				log.Println(err)
				continue
			}
		case 6:
			log.Println("Starting to download")
			urlKey := AskForUrlKey()
			validKey := make([]byte, 0)
			if AskIfPassword() {
				password, err := GetPassword()
				if err == nil {
					validKey = password
				}
			}
			data, length, err := c.DownloadFromKey(urlKey, validKey, true)
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("Downloading")
			DownloadInteraction(data, length)
		case 7:
			log.Println("Adding user")
			id := uuid.New()
			idString := id.String()
			name := AskForName()
			username := AskForUsername()
			password, err := GetPassword()
			if err != nil {
				continue
			}
			apiKey := []byte("api-key-key")
			apiKeyHash := sha512.Sum512(apiKey)
			log.Printf("Api Key hash %v", hex.EncodeToString(apiKeyHash[:]))
			userJson := users.UserJson{
				Id:   idString,
				Name: name,
				Role: users.Role{
					RootUser:       true,
					CanCreateUsers: true,
					CanAddKeys:     true,
					CanUploadData:  true,
					CanDeleteKeys:  true,
				},
				Username:     username,
				PasswordHash: password,
				ApiKey:       apiKeyHash[:],
			}
			err = c.AddUser(&userJson, RootApiKey)
			if err != nil {
				log.Println(err)
				continue
			}
		default:
			continue
		}
	}

}

func AskForBucketId() string {
	prompt := promptui.Prompt{
		Label: "Bucket Id",
	}
	res, err := prompt.Run()
	if err != nil {
		panic(res)
	}
	return res
}

func AskForKeyId() string {
	prompt := promptui.Prompt{
		Label: "Key Id",
	}
	res, err := prompt.Run()
	if err != nil {
		panic(res)
	}
	return res
}

func AskForUrlKey() string {
	prompt := promptui.Prompt{
		Label: "Url Key",
	}
	res, err := prompt.Run()
	if err != nil {
		panic(err)
	}
	return res
}

func AskIfPassword() bool {
	prompt := promptui.Select{Label: "Password?", Items: []string{"Yes", "No"}}
	_, out, err := prompt.Run()
	if err != nil {
		panic(err)
	}
	return out == "Yes"
}

func AskForName() string {
	prompt := promptui.Prompt{
		Label: "Name",
	}
	res, err := prompt.Run()
	if err != nil {
		panic(err)
	}
	return res
}

func AskForUsername() string {
	prompt := promptui.Prompt{
		Label: "Username",
	}
	res, err := prompt.Run()
	if err != nil {
		panic(err)
	}
	return res
}

func DownloadInteraction(data io.Reader, length int64) bool {
	prompt := promptui.Select{
		Label: "Select target",
		Items: []string{"File", "Stdout"},
	}
	_, choice, err := prompt.Run()
	if err != nil {
		log.Println(err)
	}
	switch choice {
	case "File":
		pathPrompt := promptui.Prompt{Label: "File path"}
		path, err := pathPrompt.Run()
		if err != nil {
			log.Fatal(err)
		}
		file, err := os.Create(path)
		if err != nil {
			log.Println(err)
			return true
		}

		bar := pb.Full.Start64(length)
		barReader := bar.NewProxyReader(data)
		_, err = file.ReadFrom(barReader)
		if err != nil {
			log.Println(err)
			return true
		}
		err = barReader.Close()
		if err != nil {
			log.Println(err)
			return true
		}
	case "Stdout":
		buf := new(strings.Builder)
		_, err := io.Copy(buf, data)
		if err != nil {
			log.Println(err)
			return true
		}
		fmt.Println(buf.String())
	}
	return false
}

func GetPassword() ([]byte, error) {
	prompt := promptui.Prompt{Label: "Enter passkey"}
	passwordString, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	passwordBytes := []byte(passwordString)
	salt := make([]byte, 128)
	key := argon2.IDKey(passwordBytes, salt, 1, 64*1024, 4, sha512.Size)
	log.Printf("Hex: %v\n", hex.EncodeToString(key))
	log.Printf("Base64: %v\n", base64.RawURLEncoding.EncodeToString(key))
	return key, nil
}
