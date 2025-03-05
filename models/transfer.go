package models

import "time"

// Model representing the "transfers" table
type Transfer struct {
	ID             int       `json:"id"`
	UserID         string    `json:"user_id"`
	GuestvoucherID int       `json:"guestvoucher_id"`
	FileCount      int       `json:"file_count"`
	TotalByteSize  int       `json:"total_byte_size"`
	Subject        string    `json:"subject"`
	Message        string    `json:"message"`
	DownloadCount  int       `json:"download_count"`
	ExpiryDate     time.Time `json:"expiry_date"`
	CreationDate   time.Time `json:"creation_date"`
}
