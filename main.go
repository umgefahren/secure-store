package main

import (
	"fmt"
	"github.com/gin-gonic/autotls"
	"log"
	"os"
	"secure-store/metadata"
	"secure-store/storage"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := storage.New()
	m := metadata.NewMemoryStore()
	r := NewRouter(s, m)

	domainsString := os.Getenv("DOMAINS")
	if domainsString == "" {
		log.Fatal(r.Run(fmt.Sprintf("0.0.0.0:%v", port)))
	} else {
		domains := strings.Split(domainsString, ",")
		domainsFiltered := make([]string, 0)
		for _, d := range domains {
			if d != "" {
				domainsFiltered = append(domainsFiltered, d)
			}
		}
		log.Fatal(autotls.Run(r, domainsFiltered...))
	}
}
