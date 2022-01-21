package main

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/manifoldco/promptui"
	"io"
	"log"
	"os"
	"secure-store/client"
	"strings"
)

func main() {
	c := client.NewClient("http://68.183.213.89:8080")
	items := []string{"Create Bucket", "Read", "Write", "Delete", "DeleteBucket", "Exit"}
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
		op, _, err := prompt.Run()
		if err != nil {
			log.Fatal(err)
		}
		if op == 5 {
			break
		}
		bucket := AskForBucketId()
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
					continue
				}

				bar := pb.Full.Start64(length)
				barReader := bar.NewProxyReader(data)
				_, err = file.ReadFrom(barReader)
				if err != nil {
					log.Println(err)
					continue
				}
				err = barReader.Close()
				if err != nil {
					log.Println(err)
					continue
				}
			case "Stdout":
				buf := new(strings.Builder)
				_, err := io.Copy(buf, data)
				if err != nil {
					log.Println(err)
					continue
				}
				fmt.Println(buf.String())
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

				err = c.Upload(bucket, key, barReader, name)
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
				err = c.Upload(bucket, key, contentReader, filename)
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
