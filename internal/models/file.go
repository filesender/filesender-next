// Package models contains data models
package models

import (
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"codeberg.org/filesender/filesender-next/internal/utils"
)

// File model representing metadata
type File struct {
	DownloadCount int `json:"download_count"`
	ByteSize      int `json:"byte_size"`
}

// Create function creates a new file metadata file representing a file
func (file *File) Create(userID string, transferID string, fileName string) error {
	uploadsPath := path.Join(os.Getenv("STATE_DIRECTORY"), "uploads")

	err := utils.WriteDataToFile(file, filepath.Join(uploadsPath, userID, transferID, fileName+".meta"))
	if err != nil {
		slog.Error("Failed writing file data", "error", err)
	}

	return err
}

// GetFileFromName creates new File object from given user ID, transfer ID, and file name
// Errors when file does not exist
func GetFileFromName(userID string, transferID string, fileName string) (File, error) {
	uploadsPath := path.Join(os.Getenv("STATE_DIRECTORY"), "uploads")
	filePath := filepath.Join(uploadsPath, userID, transferID, fileName+".meta")
	var file File

	err := utils.ReadDataFromFile(filePath, &file)
	if err != nil {
		slog.Error("Failed reading file: "+filePath, "error", err)
	}

	return file, err
}
