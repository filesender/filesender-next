package config

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// Test the initialisation of a random database & migrations
func TestInitDB(t *testing.T) {
	testDBPath := "test.db"

	db, err := InitDB(testDBPath)
	if err != nil {
		t.Errorf("Initialisation of database failed: %v", err)
	}

	db.Close()
	os.Remove(testDBPath)
}

// Test the initialisation of a random database & migrations
// Also runs migrations again (to make sure that doesn't arise any new issues)
func TestMigrationsBeingRanTwice(t *testing.T) {
	testDBPath := "test.db"

	db, err := InitDB(testDBPath)
	if err != nil {
		t.Errorf("Initialisation of database failed: %v", err)
	}

	err = runMigrations(db)
	if err != nil {
		t.Errorf("Running migrations again failed: %v", err)
	}

	db.Close()
	os.Remove(testDBPath)
}
