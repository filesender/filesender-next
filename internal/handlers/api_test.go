package handlers_test

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"codeberg.org/filesender/filesender-next/internal/crypto"
	"codeberg.org/filesender/filesender-next/internal/handlers"
	"codeberg.org/filesender/filesender-next/internal/models"
)

func TestUploadAPIHandler(t *testing.T) {
	handler := handlers.UploadAPI(10 * 1024 * 1024) // 10 MB limit

	tempDir, err := os.MkdirTemp("", "test_uploads")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed closing file %v", err)
		}
	}()

	err = os.Setenv("STATE_DIRECTORY", tempDir)
	if err != nil {
		t.Fatalf("Failed setting env var: %v", err)
	}

	// Hash "dev" for test use
	hashedID, err := crypto.HashToBase64("dev")
	if err != nil {
		t.Fatalf("Failed hashing dummy user ID: %v", err)
	}

	t.Run("Upload with invalid transfer ID", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("transfer-id", "invalid")
		_ = writer.Close()

		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})

	t.Run("Upload with valid transfer ID but no file", func(t *testing.T) {
		transfer := models.Transfer{UserID: hashedID}
		if err := transfer.Create(); err != nil {
			t.Fatalf("Failed creating new transfer: %v", err)
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("transfer-id", transfer.ID)
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
		transfer := models.Transfer{UserID: hashedID}
		if err := transfer.Create(); err != nil {
			t.Fatalf("Failed creating new transfer: %v", err)
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("transfer-id", transfer.ID)
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

		uploadedPath := path.Join(tempDir, transfer.UserID, transfer.ID, "testfile.txt")
		if _, err := os.Stat(uploadedPath); errors.Is(err, os.ErrNotExist) {
			t.Errorf("Expected uploaded file to exist: %s", uploadedPath)
		}

		metaPath := uploadedPath + ".meta"
		if _, err := os.Stat(metaPath); errors.Is(err, os.ErrNotExist) {
			t.Errorf("Expected meta file to exist: %s", metaPath)
		}
	})

	t.Run("New transfer created when no transfer-id is given", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("expiry-date", time.Now().Add(24*time.Hour).Format("2006-02-03"))
		part, _ := writer.CreateFormFile("file", "newfile.txt")
		_, _ = part.Write([]byte("Temporary file content."))
		_ = writer.Close()

		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusSeeOther {
			t.Errorf("Expected status %d for redirect, got %d", http.StatusSeeOther, resp.Code)
		}
	})
}
