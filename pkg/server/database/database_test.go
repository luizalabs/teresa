package database

import (
	"testing"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestNew(t *testing.T) {
	db, err := New(&Config{Database: ":memory:"})
	if err != nil {
		t.Fatal("error trying to create a new database connection:", err)
	}
	if db == nil {
		t.Fatal("expected a instance of gorm.DB, got nil")
	}
	if err = db.DB().Ping(); err != nil {
		t.Error("Error on make a ping to in memory database:", err)
	}
}
