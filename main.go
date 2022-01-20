package main

import (
	"fmt"
	"github.com/gin-gonic/autotls"
	"github.com/google/uuid"
	"log"
	"os"
	"secure-store/metadata"
	"secure-store/security"
	"secure-store/storage"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dataDir := os.Getenv("DATADIR")
	if dataDir == "" {
		dataDir = os.TempDir() + "data/" + uuid.NewString()
	}

	s, err := storage.NewFsStorage(dataDir)
	if err != nil {
		log.Fatal(err)
	}
	m := metadata.NewMemoryStore()
	sec := security.NewMemorySecurityStore()
	compound := CompoundStore{
		metadata: m,
		security: sec,
		storage:  s,
	}
	r := NewRouter(&compound)

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
