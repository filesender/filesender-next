package handlers_test

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/filesender/filesender-next/internal/handlers"
	"codeberg.org/filesender/filesender-next/internal/models"
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

func TestFileUpload_Success(t *testing.T) {
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

	fileContent := "This is a test file"
	testFile, cleanup, err := createMultipartFile(fileContent)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	fileMeta := models.File{
		ByteSize: int64(len(fileContent)),
	}

	err = handlers.FileUpload(tempDir, "user456", "test123", fileMeta, testFile)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	expectedPath := filepath.Join(tempDir, "user456", "test123.bin")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected file to exist at %s", expectedPath)
	}
}

func TestFileUpload_InvalidDirectory(t *testing.T) {
	testFile, cleanup, err := createMultipartFile("test")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	fileMeta := models.File{
		ByteSize: -10,
	}

	err = handlers.FileUpload("/invalid/directory/should/fail", "----------", "doesn't matter", fileMeta, testFile)
	if err == nil {
		t.Fatal("Expected error due to invalid directory, got nil")
	}
}

func TestFileUpload_CopyFailure(t *testing.T) {
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

	// Simulate a closed file (io.Copy will fail)
	fakeFile, _, err := createMultipartFile("test")
	if err != nil {
		t.Fatal(err)
	}
	err = fakeFile.Close() // manually close it to trigger failure
	if err != nil {
		t.Fatalf("Couldn't close file: %v", err)
	}

	fileMeta := models.File{
		ByteSize: 4,
	}

	err = handlers.FileUpload(tempDir, "user123", "testfail", fileMeta, fakeFile)
	if err == nil {
		t.Fatal("Expected error due to file copy failure, got nil")
	}
}
