package handlers_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/crypto"
	"codeberg.org/filesender/filesender-next/internal/handlers"
)

func TestUploadAPIHandler(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_uploads")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed closing file %v", err)
		}
	}()

	handler := handlers.UploadAPI("/", &auth.DummyAuth{}, tempDir, 10*1024*1024) // 10 MB limit

	// Hash "dev" for test use
	hashedID, err := crypto.HashToBase64("dev")
	if err != nil {
		t.Fatalf("Failed hashing dummy user ID: %v", err)
	}

	t.Run("Upload with no file", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.Close()

		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})

	t.Run("Successful file upload", func(t *testing.T) {
		datePlusTwo := time.Now().AddDate(0, 0, 2).Format("2006-01-02")

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("expiry-date", datePlusTwo)
		part, _ := writer.CreateFormFile("file", "testfile.txt")
		_, _ = part.Write([]byte("This is a test file."))
		_ = writer.Close()

		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusSeeOther {
			t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.Code)
		}

		files, err := os.ReadDir(path.Join(tempDir, hashedID))
		if err != nil {
			t.Fatalf("Failed to read directory: %v", err)
		}

		if len(files) != 2 {
			t.Errorf("Expected 2 files in directory, found %d", len(files))
		}
	})
}
