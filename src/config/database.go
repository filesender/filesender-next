package config

import (
	"database/sql"
	"log"
)

// Initializes a SQLite database at the given path.
// Opens the database connection and applies "migrations".
func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	err = runMigrations(db)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Database connected and migrated successfully.")
	return db, nil
}

// Runs "migrations" on given database
func runMigrations(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS transfers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT,
			guestvoucher_id INTEGER,
			file_count INTEGER DEFAULT 0,
			total_byte_size BIGINT DEFAULT 0,
			subject TEXT,
			message TEXT,
			download_count INTEGER DEFAULT 0,
			expiry_date DATETIME,
			creation_date DATETIME DEFAULT (CURRENT_TIMESTAMP)
		);

		CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			transfer_id INTEGER NOT NULL,
			file_name STRING NOT NULL,
			file_byte_size BIGINT DEFAULT 0,
			download_count INTEGER DEFAULT 0,
			FOREIGN KEY (transfer_id) REFERENCES transfers(id) ON DELETE CASCADE
		)
	`

	_, err := db.Exec(query)
	return err
}
