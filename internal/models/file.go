package models

import (
	"encoding/json"
	"os"
	"path"
)

// Model representing the "files" metadata
type File struct {
	DownloadCount int `json:"download_count"`
}

func (file *File) Create(uploadDest string, fileName string) error {
	metaJSON, err := json.Marshal(file)
	if err != nil {
		return err
	}

	metaFile, err := os.Create(path.Join(uploadDest, fileName+".meta"))
	if err != nil {
		return err
	}
	defer metaFile.Close()

	_, err = metaFile.Write(metaJSON)
	if err != nil {
		return err
	}

	return nil
}
