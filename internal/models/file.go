package models

// Model representing the "files" table
type File struct {
	ID            int    `json:"id"`
	TransferID    int    `json:"transfer_id"`
	FileName      string `json:"file_name"`
	FileByteSize  int    `json:"file_byte_size"`
	DownloadCount int    `json:"download_count"`
}
