// Package models contains data models
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

// File model representing metadata
type File struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	FileNames     []string  `json:"file_names"`
	FileCount     int       `json:"file_count"`
	DownloadCount int       `json:"download_count"`
	ByteSize      int       `json:"byte_size"`
	Path          string    `json:"path"`
	ExpiryDate    time.Time `json:"expiry_date"`
	CreationDate  time.Time `json:"creation_date"`
}

// Create function creates a new meta file & sets the creation date
func (file *File) Create() error {
	file.CreationDate = time.Now().UTC().Round(time.Second)
	return file.Save()
}

// Save the current File meta state to disk
func (file *File) Save() error {
	err := json.WriteDataToFile(file, filepath.Join(os.Getenv("STATE_DIRECTORY"), file.UserID, file.ID+".tar.meta"))
	if err != nil {
		slog.Error("Failed writing meta file", "error", err)
		return err
	}

	return nil
}

// FileExists checks if the file exists
func FileExists(userID string, fileID string) (bool, error) {
	filePath := filepath.Join(os.Getenv("STATE_DIRECTORY"), userID, fileID+".tar.meta")

	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

// ValidateFile checks if file ID is valid & if the file exists
func ValidateFile(userID, fileID string) error {
	err := id.Validate(fileID)
	if err != nil {
		slog.Error("Invalid file ID", "error", err)
		return fmt.Errorf("file ID is invalid")
	}

	exists, err := FileExists(userID, fileID)
	if err != nil {
		slog.Error("File not found", "error", err)
		return fmt.Errorf("could not find file")
	}

	if !exists {
		slog.Error("File does not exist!", "userID", userID, "fileID", fileID)
		return fmt.Errorf("could not find file")
	}

	return nil
}

// GetFileFromIDs creates new File object from given user ID & file ID
// Errors when file does not exist
func GetFileFromIDs(userID string, fileID string) (File, error) {
	filePath := filepath.Join(os.Getenv("STATE_DIRECTORY"), userID, fileID+".tar")
	var file File

	err := json.ReadDataFromFile(filePath+".meta", &file)
	if err != nil {
		slog.Error("Failed reading file: "+filePath+".meta", "error", err)
	}

	file.Path = filePath
	return file, err
}
