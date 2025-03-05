package config

import (
	"io"
	"os"
	"testing"
)

// Copy env.example to .env
// Test loading of the .env file
func TestEnvReading(t *testing.T) {
	_, exists := os.LookupEnv("DATABASE_PATH")
	if exists {
		t.Errorf("Database path is already set in environment variables")
	}

	err := copyFile("../env.example", ".env")
	if err != nil {
		t.Errorf("Failed copying env.example to .env: %v", err)
	}

	LoadEnv()

	_, exists = os.LookupEnv("DATABASE_PATH")
	if !exists {
		t.Errorf("Failed setting database path through .env file")
	}

	os.Remove(".env")
}

func copyFile(src, dest string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
