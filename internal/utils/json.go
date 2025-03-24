package utils

import (
	"encoding/json"
	"os"
)

// Receives any data object and writes to path
func WriteDataFromFile(data any, path string) error {
	JSONData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(JSONData)
	if err != nil {
		return err
	}

	return nil
}

// Reads data from file into object
func ReadDataFromFile(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}
