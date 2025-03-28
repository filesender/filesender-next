// Package models contains data models
package models

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"codeberg.org/filesender/filesender-next/internal/json"
)

// File model representing metadata
type File struct {
	DownloadCount int `json:"download_count"`
	ByteSize      int `json:"byte_size"`
	Path          string
}

// Create function creates a new file metadata file representing a file
func (file *File) Create(userID string, transferID string, fileName string) error {
	err := json.WriteDataToFile(file, filepath.Join(os.Getenv("STATE_DIRECTORY"), userID, transferID, fileName+".meta"))
	if err != nil {
		slog.Error("Failed writing file data", "error", err)
	}

	return err
}

// FileExists checks if the file exists
func FileExists(userID string, transferID string, fileName string) (bool, error) {
	filePath := filepath.Join(os.Getenv("STATE_DIRECTORY"), userID, transferID, fileName+".meta")

	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

// GetFileFromName creates new File object from given user ID, transfer ID, and file name
// Errors when file does not exist
func GetFileFromName(userID string, transferID string, fileName string) (File, error) {
	filePath := filepath.Join(os.Getenv("STATE_DIRECTORY"), userID, transferID, fileName)
	var file File

	err := json.ReadDataFromFile(filePath+".meta", &file)
	if err != nil {
		slog.Error("Failed reading file: "+filePath+".meta", "error", err)
	}

	file.Path = filePath
	return file, err
}
