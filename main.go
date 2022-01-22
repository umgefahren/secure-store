package main

import (
	"fmt"
	"github.com/gin-gonic/autotls"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"log"
	"os"
	"secure-store/access"
	"secure-store/metadata"
	"secure-store/security"
	"secure-store/storage"
	"strings"
)

func main() {
	fmt.Println("Welcome to Secure-Store v0.0.1 👋")
	fmt.Println("I will keep your files secure and accessible 🔒")
	fmt.Println(" ❌  Don't use this software in production!! ❌  ")
	fmt.Println("@umgefahren")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logrus.WithField("port", port).Info("Starting server with port parameter")

	dataDir := os.Getenv("DATADIR")
	if dataDir == "" {
		dataDir = os.TempDir() + "data/" + uuid.NewString()
	}

	logrus.WithField("Data Directory", dataDir).Info("Starting server with data directory paramter")
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
	db := sqlite.Open("test.db")
	a, err := access.NewSQLStore(db)
	if err != nil {
		log.Fatal(err)
	}
	r := NewRouter(&compound, a)

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
