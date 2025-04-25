// Package models contains data models
package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"codeberg.org/filesender/filesender-next/internal/id"
)

// File model representing metadata
type File struct {
	ByteSize          int64     `json:"byte_size"`
	EncryptedFileName string    `json:"encrypted_file_name"`
	Chunked           bool      `json:"chunked"`
	Partial           bool      `json:"partial"`
	Chunks            []string  `json:"chunks"`
	ChunkSize         int64     `json:"chunk_size"`
	ExpiryDate        time.Time `json:"expiry_date"`
	CreationDate      time.Time `json:"creation_date"`
}

// Create function creates a new meta file & sets the creation date
func (file *File) Create(stateDir string, userID string, fileID string) error {
	file.CreationDate = time.Now().UTC().Round(time.Second)
	return file.Save(stateDir, userID, fileID)
}

// Save the current File meta state to disk
func (file *File) Save(stateDir string, userID string, fileID string) error {
	JSONData, err := json.Marshal(file)
	if err != nil {
		slog.Error("Failed marshalling data", "error", err)
		return err
	}

	metaPath := filepath.Join(stateDir, userID, fileID+".meta")
	f, err := os.OpenFile(metaPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		slog.Error("Failed opening file", "error", err, "path", metaPath)
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			slog.Error("Failed closing file", "error", err, "path", metaPath)
		}
	}()

	_, err = f.Write(JSONData)
	if err != nil {
		slog.Error("Failed writing data", "error", err, "path", metaPath)
		return err
	}

	return nil
}

// FileExists checks if the file exists
func FileExists(stateDir string, userID string, fileID string) (bool, error) {
	filePath := filepath.Join(stateDir, userID, fileID+".meta")

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
func ValidateFile(stateDir string, userID string, fileID string) error {
	err := id.Validate(fileID)
	if err != nil {
		slog.Error("Invalid file ID", "error", err)
		return fmt.Errorf("file ID is invalid")
	}

	exists, err := FileExists(stateDir, userID, fileID)
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
func GetFileFromIDs(stateDir string, userID string, fileID string) (File, error) {
	metaPath := filepath.Join(stateDir, userID, fileID+".meta")
	var file File

	data, err := os.ReadFile(metaPath)
	if err != nil {
		slog.Error("Failed reading file", "error", err, "path", metaPath)
		return file, err
	}

	err = json.Unmarshal(data, &file)
	if err != nil {
		slog.Error("Failed unmarshalling meta file", "error", err, "path", metaPath)
	}

	return file, err
}
