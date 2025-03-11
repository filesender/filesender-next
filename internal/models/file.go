package models

import (
	"database/sql"
	"log/slog"
)

// Model representing the "files" table
type File struct {
	ID            int    `json:"id"`
	TransferID    int    `json:"transfer_id"`
	FileName      string `json:"file_name"`
	FileByteSize  int    `json:"file_byte_size"`
	DownloadCount int    `json:"download_count"`
}

// Count amount of files in the database
func CountFiles(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT count(id) as c FROM files").Scan(&count)
	if err != nil {
		slog.Error("Database query error", "error", err)
		return 0, err
	}
	return count, nil
}
