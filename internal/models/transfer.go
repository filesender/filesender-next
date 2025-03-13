package models

import (
	"database/sql"
	"time"
)

// Model representing the "transfers" table
type Transfer struct {
	ID            int       `json:"id"`
	UserID        string    `json:"user_id"`
	FileCount     int       `json:"file_count"`
	TotalByteSize int       `json:"total_byte_size"`
	Subject       string    `json:"subject"`
	Message       string    `json:"message"`
	DownloadCount int       `json:"download_count"`
	ExpiryDate    time.Time `json:"expiry_date"`
	CreationDate  time.Time `json:"creation_date"`
}

func CreateTransfer(db *sql.DB, transfer Transfer) (*Transfer, error) {
	query := `
		INSERT INTO transfers (
			user_id, guestvoucher_id, file_count, total_byte_size, subject, message, download_count, expiry_date
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id;
	`

	result, err := db.Exec(query,
		transfer.UserID,
		transfer.FileCount,
		transfer.TotalByteSize,
		transfer.Subject,
		transfer.Message,
		transfer.DownloadCount,
		transfer.ExpiryDate,
	)
	if err != nil {
		return nil, err
	}
	transfer.CreationDate = time.Now().UTC().Round(time.Second)

	transferID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	transfer.ID = int(transferID)

	return &transfer, nil
}
