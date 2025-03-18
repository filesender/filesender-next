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

// Creates new transfer record in database based on transfer struct
func (transfer *Transfer) Create(db *sql.DB) error {
	query := `
		INSERT INTO transfers (
			user_id, file_count, total_byte_size, subject, message, download_count, expiry_date
		) VALUES (?, ?, ?, ?, ?, ?, ?)
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
		return err
	}
	transfer.CreationDate = time.Now().UTC().Round(time.Second)

	transferID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	transfer.ID = int(transferID)

	return nil
}

// Adds data of a new file to transfer
func (transfer *Transfer) NewFile(db *sql.DB, byteSize int) error {
	query := "UPDATE transfers SET file_count = file_count + 1, total_byte_size = total_byte_size + ? WHERE id = ?"
	_, err := db.Exec(query, byteSize, transfer.ID)
	if err != nil {
		return err
	}

	transfer.TotalByteSize += byteSize
	transfer.FileCount++
	return nil
}

// Removes data of a file from transfer
func (transfer *Transfer) RemoveFile(db *sql.DB, byteSize int) error {
	query := "UPDATE transfers SET file_count = file_count - 1, total_byte_size = total_byte_size - ? WHERE id = ?"
	_, err := db.Exec(query, byteSize, transfer.ID)
	if err != nil {
		return err
	}

	transfer.TotalByteSize -= byteSize
	transfer.FileCount--
	return nil
}

// Gets transfer based on ID
func GetTransferFromID(db *sql.DB, id int) (Transfer, error) {
	transfer := Transfer{}
	err := db.QueryRow("SELECT * FROM transfers WHERE id = ?", id).Scan(
		&transfer.ID,
		&transfer.UserID,
		&transfer.FileCount,
		&transfer.TotalByteSize,
		&transfer.Subject,
		&transfer.Message,
		&transfer.DownloadCount,
		&transfer.ExpiryDate,
		&transfer.CreationDate,
	)
	if err != nil {
		return transfer, err
	}

	return transfer, nil
}
