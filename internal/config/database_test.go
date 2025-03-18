package config

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// Helper function
func createDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Test the initialisation of a random database & migrations
func TestInitDB(t *testing.T) {
	testDBPath := "test.db"
	db, err := createDB(testDBPath)
	if err != nil {
		t.Errorf("Initialisation of database failed: %v", err)
	}

	err = InitDB(db)
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
	db, err := createDB(testDBPath)
	if err != nil {
		t.Errorf("Initialisation of database failed: %v", err)
	}

	err = InitDB(db)
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
