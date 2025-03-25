// Package models contains data models
package models

import (
	"encoding/json"
	"log/slog"
	"os"
	"path"
)

// File model representing metadata
type File struct {
	DownloadCount int `json:"download_count"`
}

// Create function creates a new file metadata file representing a file
func (file *File) Create(uploadDest string, fileName string) error {
	metaJSON, err := json.Marshal(file)
	if err != nil {
		return err
	}

	metaFile, err := os.Create(path.Join(uploadDest, fileName+".meta"))
	if err != nil {
		return err
	}
	defer func() {
		if err := metaFile.Close(); err != nil {
			slog.Error("Failed closing file", "error", err)
		}
	}()

	_, err = metaFile.Write(metaJSON)
	if err != nil {
		return err
	}

	return nil
}
