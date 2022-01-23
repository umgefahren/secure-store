package access

import (
	"gorm.io/driver/sqlite"
	"testing"
)

func TestAddAndDeleteSQL(t *testing.T) {

	dbSqlite := sqlite.Open("file::memory:?cache=shared")
	db, err := NewSQLStore(dbSqlite)
	if err != nil {
		t.Fatal(err)
	}
	AddAndDeleteTest(t, db)
}

func TestAddAndAccessSQL(t *testing.T) {

	dbSqlite := sqlite.Open("file::memory:?cache=shared")
	db, err := NewSQLStore(dbSqlite)
	if err != nil {
		t.Fatal(err)
	}
	AddAndAccessTest(t, db)
}
