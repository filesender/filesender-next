package handlers_test

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/filesender/filesender-next/internal/handlers"
)

func createMultipartFile(content string) (multipart.File, func(), error) {
	tmpFile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		return nil, nil, err
	}

	_, err = tmpFile.Write([]byte(content))
	if err != nil {
		e := tmpFile.Close()
		if e != nil {
			return nil, nil, e
		}
		return nil, nil, err
	}

	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		e := tmpFile.Close()
		if e != nil {
			return nil, nil, e
		}
		return nil, nil, err
	}

	cleanup := func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}

	return tmpFile, cleanup, nil
}

func TestFileUploadHandler(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_uploads")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Errorf("Failed deleting temp dir: %v", err)
		}
	}()

	t.Run("State directory does not exist", func(t *testing.T) {
		testFile, cleanup, err := createMultipartFile("")
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()

		var pathErr *os.PathError
		err = handlers.FileUpload("/hello/world", "user", "file", testFile)
		if err == nil {
			t.Errorf("Expected file upload to result nil, got %v", err)
		} else if !errors.As(err, &pathErr) {
			t.Errorf("Expected error to be os.PathError, got %v", err)
		}
	})

	t.Run("Successful upload", func(t *testing.T) {
		fileContent := "test test"
		testFile, cleanup, err := createMultipartFile(fileContent)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()

		err = handlers.FileUpload(tempDir, "user456", "test123", testFile)
		if err != nil {
			t.Fatalf("Expected success, got error: %v", err)
		}

		expectedPath := filepath.Join(tempDir, "user456", "test123")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("Expected file to exist at %s", expectedPath)
		}
	})

	t.Run("Copy fail", func(t *testing.T) {
		fakeFile, _, err := createMultipartFile("test")
		if err != nil {
			t.Fatal(err)
		}
		err = fakeFile.Close() // manually close it to trigger failure
		if err != nil {
			t.Fatalf("Couldn't close file: %v", err)
		}

		err = handlers.FileUpload(tempDir, "user123", "testfail", fakeFile)
		if err == nil {
			t.Fatal("Expected error due to file copy failure, got nil")
		} else if !strings.Contains(err.Error(), "file already closed") {
			t.Fatalf("Expected \"file already closed\" error, instead got: %v", err)
		}
	})
}

func TestPartialFileUpload(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_uploads")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Errorf("Failed deleting temp dir: %v", err)
		}
	}()

	// Create user directory for user "user"
	testFile, cleanup, err := createMultipartFile("Hello, world!")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	handlers.FileUpload(tempDir, "user", "file", testFile)

	t.Run("State directory does not exist", func(t *testing.T) {
		testFile, cleanup, err := createMultipartFile("")
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()

		var pathErr *os.PathError
		_, err = handlers.PartialFileUpload("/hello/world", "user", "file", testFile, 0)
		if err == nil {
			t.Errorf("Expected file upload to result nil, got %v", err)
		} else if !errors.As(err, &pathErr) {
			t.Errorf("Expected error to be os.PathError, got %v", err)
		}
	})

	t.Run("File does not exist", func(t *testing.T) {
		testFile, cleanup, err := createMultipartFile("")
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()

		var pathErr *os.PathError
		_, err = handlers.PartialFileUpload(tempDir, "user", "file_fail", testFile, 0)
		if err == nil {
			t.Errorf("Expected file upload to error, got %v", err)
		} else if !errors.As(err, &pathErr) {
			t.Errorf("Expected error to be os.PathError, got %v", err)
		}
	})

	t.Run("Copy fail", func(t *testing.T) {
		fakeFile, _, err := createMultipartFile("test")
		if err != nil {
			t.Fatal(err)
		}
		err = fakeFile.Close() // manually close it to trigger failure
		if err != nil {
			t.Fatalf("Couldn't close file: %v", err)
		}

		err = handlers.FileUpload(tempDir, "user", "file", fakeFile)
		if err == nil {
			t.Fatal("Expected error due to file copy failure, got nil")
		} else if !strings.Contains(err.Error(), "file already closed") {
			t.Fatalf("Expected \"file already closed\" error, instead got: %v", err)
		}
	})

	t.Run("Overwrite existing data", func(t *testing.T) {
		testFile, cleanup, err := createMultipartFile("Hello, world!")
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()

		totalBytes, err := handlers.PartialFileUpload(tempDir, "user", "file", testFile, 0)
		if err != nil {
			t.Errorf("Did not expect error, got %v", err)
		}
		if totalBytes != 13 {
			t.Errorf("Expected total bytes in file to be 13, got %d", totalBytes)
		}
	})

	t.Run("Success", func(t *testing.T) {
		testFile, cleanup, err := createMultipartFile("Hello, world!")
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()

		totalBytes, err := handlers.PartialFileUpload(tempDir, "user", "file", testFile, 13)
		if err != nil {
			t.Errorf("Did not expect error, got %v", err)
		}
		if totalBytes != 26 {
			t.Errorf("Expected total bytes in file to be 26, got %d", totalBytes)
		}
	})
}
