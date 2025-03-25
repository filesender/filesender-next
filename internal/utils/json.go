// Package utils contains small utility functions
package utils

import (
	"encoding/json"
	"log/slog"
	"os"
)

// WriteDataFromFile receives any data object and writes to path
func WriteDataFromFile(data any, path string) error {
	JSONData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error("Failed closing file", "error", err)
		}
	}()

	_, err = file.Write(JSONData)
	if err != nil {
		return err
	}

	return nil
}

// ReadDataFromFile reads data from file into object
func ReadDataFromFile(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}
