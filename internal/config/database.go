package config

import (
	"database/sql"
	"log/slog"
	"os"
)

// Initializes a SQLite database at the given path.
// Opens the database connection and applies "migrations".
func InitDB(path string) (*sql.DB, error) {
	slog.Debug("Using database", "path", path)
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	err = runMigrations(db)
	if err != nil {
		slog.Error("Migration failed", "error", err)
		os.Exit(1)
	}

	slog.Debug("Database connected and migrated successfully")
	return db, nil
}

// Runs "migrations" on given database
func runMigrations(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS transfers (
			id INTEGER PRIMARY KEY,
			user_id TEXT,
			file_count INTEGER DEFAULT 0,
			total_byte_size BIGINT DEFAULT 0,
			subject TEXT,
			message TEXT,
			download_count INTEGER DEFAULT 0,
			expiry_date TIMESTAMP,
			creation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := db.Exec(query)
	return err
}
