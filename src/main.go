package main

import (
	"log"
	"os"

	"codeberg.org/filesender/filesender-next/config"

	_ "modernc.org/sqlite"
)

func main() {
	err := LoadEnv()
	if err != nil {
		log.Fatalf("Loading env file failed: %v", err)
		return
	}

	db_path, exists := os.LookupEnv("DATABASE_PATH")
	if !exists {
		log.Fatalf("Database path not set in env")
		return
	}

	db, err := config.InitDB(db_path)
	if err != nil {
		log.Fatalf("Failed initialising database: %v", err)
		return
	}

	// Placeholder
	_ = db
}
