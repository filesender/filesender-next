package models

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"codeberg.org/filesender/filesender-next/internal/id"
	"codeberg.org/filesender/filesender-next/internal/json"
)

// Transfer model representing metadata
type Transfer struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	FileNames     []string  `json:"file_names"`
	FileCount     int       `json:"file_count"`
	TotalByteSize int       `json:"total_byte_size"`
	DownloadCount int       `json:"download_count"`
	ExpiryDate    time.Time `json:"expiry_date"`
	CreationDate  time.Time `json:"creation_date"`
}

// Create function creates a new transfer folder & writes a new meta file
func (transfer *Transfer) Create() error {
	stateDir := os.Getenv("STATE_DIRECTORY")
	transferID, err := id.New()
	if err != nil {
		slog.Error("Failed getting transfer ID", "error", err)
		return err
	}
	transfer.ID = transferID

	userDir := filepath.Clean(filepath.Join(stateDir, transfer.UserID))
	err = os.MkdirAll(userDir, 0o700)
	if err != nil {
		slog.Error("Failed creating user directory", "error", err)
		return err
	}

	transfer.CreationDate = time.Now().UTC().Round(time.Second)
	return transfer.Save()
}

// Save the current Transfer state to disk
func (transfer *Transfer) Save() error {
	stateDir := os.Getenv("STATE_DIRECTORY")

	err := json.WriteDataToFile(transfer, filepath.Join(stateDir, transfer.UserID, transfer.ID+".meta"))
	if err != nil {
		slog.Error("Failed writing meta file", "error", err)
		return err
	}

	return nil
}

// NewFile adds data of a new file to transfer
func (transfer *Transfer) NewFile(fileName string, byteSize int) error {
	transfer.TotalByteSize += byteSize
	transfer.FileCount++
	transfer.FileNames = append(transfer.FileNames, fileName)

	err := transfer.Save()
	if err != nil {
		slog.Error("Failed adding new file data to transfer metadata", "error", err)
		return err
	}

	return nil
}

// ValidateTransfer checks if transfer ID is valid & if the transfer exists
func ValidateTransfer(userID, transferID string) error {
	err := id.Validate(transferID)
	if err != nil {
		slog.Error("Invalid transfer ID", "error", err)
		return fmt.Errorf("transfer ID is invalid")
	}

	err = transferExists(userID, transferID)
	if err != nil {
		slog.Error("Transfer not found", "error", err)
		return fmt.Errorf("could not find transfer")
	}

	return nil
}

func transferExists(userID string, transferID string) error {
	stateDir := os.Getenv("STATE_DIRECTORY")

	if _, err := os.Stat(filepath.Join(stateDir, userID, transferID+".meta")); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Transfer ID \"%s\" with user ID \"%s\" does not exist", transferID, userID)
	}

	return nil
}

// GetTransferFromIDs gets transfer based on user ID & transfer ID
func GetTransferFromIDs(userID string, transferID string) (Transfer, error) {
	stateDir := os.Getenv("STATE_DIRECTORY")
	var transfer Transfer

	err := json.ReadDataFromFile(filepath.Join(stateDir, userID, transferID+".meta"), &transfer)
	if err != nil {
		slog.Error("Failed reading metadata", "error", err)
		return transfer, err
	}

	return transfer, nil
}
