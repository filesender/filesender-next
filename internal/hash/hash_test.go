package hash_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/filesender/filesender-next/internal/hash"
)

func TestInit(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_uploads")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed closing file %v", err)
		}
	}()

	t.Run("State is invalid", func(t *testing.T) {
		os.Mkdir(filepath.Join(tempDir, "hmac.key"), 0755)
		defer os.Remove(filepath.Join(tempDir, "hmac.key"))

		err = hash.Init(tempDir)
		if err == nil {
			t.Errorf("Was supposed to return error, instead got nil")
		} else if !strings.Contains(err.Error(), "read key: ") {
			t.Errorf("Was supposed to test \"read key\" instead, got \"%s\"", strings.Split(err.Error(), ": ")[0])
		} else if !strings.Contains(err.Error(), ": is a directory") {
			t.Errorf("Expected error \"is a directory\", got \"%s\"", strings.Split(err.Error(), ": ")[2])
		}
	})

	t.Run("Success", func(t *testing.T) {
		err = hash.Init(tempDir)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})
}

func TestToBase64(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_uploads")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed closing file %v", err)
		}
	}()

	t.Run("Key not initialised", func(t *testing.T) {
		hash.ResetKeyForTest()
		defer hash.Init(tempDir)

		_, err = hash.ToBase64("test")
		if err == nil {
			t.Errorf("Expected error, got none")
		} else if !strings.Contains(err.Error(), "key not here") {
			t.Errorf("Expected error to contain \"key not here\", got: \"%s\"", err.Error())
		}
	})

	t.Run("Success", func(t *testing.T) {
		h, err := hash.ToBase64("Hello, world!")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if h == "" {
			t.Errorf("Expected output to not be empty string")
		}
	})
}
