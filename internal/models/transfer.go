package models

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"codeberg.org/filesender/filesender-next/internal/utils"
	"github.com/google/uuid"
)

// Model representing the "transfers" table
type Transfer struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	FileCount     int       `json:"file_count"`
	TotalByteSize int       `json:"total_byte_size"`
	Subject       string    `json:"subject"`
	Message       string    `json:"message"`
	DownloadCount int       `json:"download_count"`
	ExpiryDate    time.Time `json:"expiry_date"`
	CreationDate  time.Time `json:"creation_date"`
}

// Creates new transfer folder and writes meta file
func (transfer *Transfer) Create() error {
	stateDir := os.Getenv("STATE_DIRECTORY")
	transfer.ID = uuid.New().String()

	userDir := filepath.Clean(filepath.Join(stateDir, "uploads", transfer.UserID))
	err := os.MkdirAll(userDir, os.ModePerm)
	if err != nil {
		slog.Error("Failed creating user directory", "error", err)
		return err
	}

	transfer.CreationDate = time.Now().UTC().Round(time.Second)

	err = utils.WriteDataFromFile(transfer, filepath.Join(userDir, transfer.ID+".meta"))
	if err != nil {
		slog.Error("Failed writing meta file", "error", err)
		return err
	}

	return nil
}

func (transfer *Transfer) Update() error {
	stateDir := os.Getenv("STATE_DIRECTORY")

	err := utils.WriteDataFromFile(transfer, filepath.Join(stateDir, "uploads", transfer.UserID, transfer.ID+".meta"))
	if err != nil {
		slog.Error("Failed writing meta file", "error", err)
		return err
	}

	return nil
}

// Adds data of a new file to transfer
func (transfer *Transfer) NewFile(byteSize int) error {
	transfer.TotalByteSize += byteSize
	transfer.FileCount++

	err := transfer.Update()
	if err != nil {
		slog.Error("Failed adding new file data to transfer metadata", "error", err)
		return err
	}

	return nil
}

// Gets transfer based on user ID & transfer ID
func GetTransferFromIDs(userID string, transferID string) (Transfer, error) {
	stateDir := os.Getenv("STATE_DIRECTORY")
	var transfer Transfer

	err := utils.ReadDataFromFile(filepath.Join(stateDir, "uploads", userID, transferID+".meta"), &transfer)
	if err != nil {
		slog.Error("Failed reading metadata", "error", err)
		return transfer, err
	}

	return transfer, nil
}
