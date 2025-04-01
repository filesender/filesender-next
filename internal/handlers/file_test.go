package handlers_test

import (
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"testing"

	"codeberg.org/filesender/filesender-next/internal/handlers"
	"codeberg.org/filesender/filesender-next/internal/models"
)

// Helper function for tests
func createMultipartFile(content string) (*multipart.FileHeader, *os.File, error) {
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		return nil, nil, err
	}

	_, err = tmpFile.Write([]byte(content))
	if err != nil {
		err = tmpFile.Close()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, err
	}

	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		err = tmpFile.Close()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, err
	}

	fileHeader := &multipart.FileHeader{Filename: "testfile.txt"}

	return fileHeader, tmpFile, nil
}

func TestHandleFileUpload_Success(t *testing.T) {
	// Set up a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test_uploads")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed deleting temp directory: %v", err)
		}
	}()

	err = os.Setenv("STATE_DIRECTORY", tempDir)
	if err != nil {
		t.Fatalf("Failed setting env var: %v", err)
	}

	transfer := models.Transfer{
		ID:     "testing123",
		UserID: "weeb",
	}
	err = os.MkdirAll(filepath.Join(tempDir, transfer.UserID), os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create user uploads directory: %v", err)
	}

	// Create a fake file
	fileHeader, file, err := createMultipartFile("This is a test file.")
	if err != nil {
		t.Fatalf("Failed to create multipart file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed closing file: %v", err)
		}
	}()

	err = handlers.FileUpload(transfer, file, fileHeader)
	if err != nil {
		t.Fatalf("HandleFileUpload failed: %v", err)
	}

	// Verify that the file was created
	uploadDest := path.Join(tempDir, transfer.UserID, transfer.ID, fileHeader.Filename)
	if _, err := os.Stat(uploadDest); os.IsNotExist(err) {
		t.Fatalf("Expected file %s to exist, but it does not", uploadDest)
	}
}

func TestHandleFileUpload_Failure_InvalidDirectory(t *testing.T) {
	// Set an invalid directory
	err := os.Setenv("STATE_DIRECTORY", "/invalid/directory")
	if err != nil {
		t.Fatalf("Failed setting env var: %v", err)
	}

	transfer := models.Transfer{
		ID:     "testing123",
		UserID: "weeb",
	}

	fileHeader, file, err := createMultipartFile("Test file.")
	if err != nil {
		t.Fatalf("Failed to create multipart file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed closing file: %v", err)
		}
	}()

	// Expect an error
	err = handlers.FileUpload(transfer, file, fileHeader)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

func TestHandleFileUpload_Failure_FileCreation(t *testing.T) {
	// Set a temporary directory
	tempDir, err := os.MkdirTemp("", "test_uploads")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed deleting temp directory: %v", err)
		}
	}()

	err = os.Setenv("STATE_DIRECTORY", tempDir)
	if err != nil {
		t.Fatalf("Failed setting env var: %v", err)
	}

	transfer := models.Transfer{
		ID:     "testing123",
		UserID: "weeb",
	}
	err = os.MkdirAll(filepath.Join(tempDir, transfer.UserID), os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create user uploads directory: %v", err)
	}

	// Create a fake file
	fileHeader, file, err := createMultipartFile("File content.")
	if err != nil {
		t.Fatalf("Failed to create multipart file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Failed closing file: %v", err)
		}
	}()

	// Set empty filename to cause failure
	fileHeader.Filename = ""

	// Expect an error
	err = handlers.FileUpload(transfer, file, fileHeader)
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}
