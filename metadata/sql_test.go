package metadata

import (
	"gorm.io/driver/sqlite"
	"testing"
)

func TestBuckets(t *testing.T) {
	dbSqlite := sqlite.Open("file::memory:?cache=shared")
	db, err := NewSQLStore(dbSqlite)
	if err != nil {
		t.Fatal(err)
	}
	BucketTest(t, db)
}

func TestWriteAndRead(t *testing.T) {
	dbSqlite := sqlite.Open("file::memory:?cache=shared")
	db, err := NewSQLStore(dbSqlite)
	if err != nil {
		t.Fatal(err)
	}
	ReadAndWriteTest(t, db)
}

func TestDelete(t *testing.T) {
	dbSqlite := sqlite.Open("file::memory:?cache=shared")
	db, err := NewSQLStore(dbSqlite)
	if err != nil {
		t.Fatal(err)
	}
	DeleteTest(t, db)
}
