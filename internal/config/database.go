package config

import (
	"database/sql"
	"log/slog"
)

// Initializes a SQLite database at the given path.
// Opens the database connection and applies "migrations".
func InitDB(db *sql.DB) error {
	err := runMigrations(db)
	if err != nil {
		slog.Error("Migration failed", "error", err)
		return err
	}

	slog.Debug("Database connected and migrated successfully")
	return nil
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
