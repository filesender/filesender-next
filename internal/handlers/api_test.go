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

	"codeberg.org/filesender/filesender-next/internal/handlers"
	"codeberg.org/filesender/filesender-next/internal/models"
)

func TestUploadAPIHandler(t *testing.T) {
	handler := handlers.UploadAPI(10 * 1024 * 1024) // 10 MB limit

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
		UserID: "dummy_session",
	}
	err = transfer.Create()
	if err != nil {
		t.Fatalf("Failed to create temporary transfer: %v", err)
	}

	t.Run("Upload without authentication", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/upload", nil)
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
		}
	})

	t.Run("Upload with invalid transfer ID", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("transfer_id", "invalid")
		err = writer.Close()
		if err != nil {
			t.Fatalf("Failed closing HTTP body writer: %v", err)
		}

		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("REMOTE_USER", "dummy_session")
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.RemoteAddr = "127.0.0.1:12345"
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})

	t.Run("Upload with valid transfer ID but no file", func(t *testing.T) {
		transfer := models.Transfer{
			UserID: "f4ZHx-yLnGfBdbhjzGtAP7hPora2QMFl6qcEdt1hJgk", // hashed "dummy_session"
		}
		err = transfer.Create()
		if err != nil {
			t.Fatalf("Failed creating new transfer: %v", err)
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("transfer_id", transfer.ID)
		err = writer.Close()
		if err != nil {
			t.Fatalf("Failed closing HTTP body writer: %v", err)
		}

		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("REMOTE_USER", "dummy_session")
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.RemoteAddr = "127.0.0.1:12345"
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Successful file upload", func(t *testing.T) {
		transfer := models.Transfer{
			UserID: "f4ZHx-yLnGfBdbhjzGtAP7hPora2QMFl6qcEdt1hJgk", // hashed "dummy_session"
		}
		err = transfer.Create()
		if err != nil {
			t.Fatalf("Failed creating new transfer: %v", err)
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("transfer_id", transfer.ID)
		part, _ := writer.CreateFormFile("file", "testfile.txt")
		_, err = part.Write([]byte("This is a test file."))
		if err != nil {
			t.Errorf("Failed creating file contents")
		}
		err = writer.Close()
		if err != nil {
			t.Fatalf("Failed closing HTTP body writer: %v", err)
		}

		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("REMOTE_USER", "dummy_session")
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.RemoteAddr = "127.0.0.1:12345"
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.Code)
		}

		if _, err := os.Stat(path.Join(tempDir, "uploads", transfer.UserID, transfer.ID, "testfile.txt")); errors.Is(err, os.ErrNotExist) {
			t.Errorf("Expected newly uploaded file to exist, it doesn't")
		}

		if _, err := os.Stat(path.Join(tempDir, "uploads", transfer.UserID, transfer.ID, "testfile.txt.meta")); errors.Is(err, os.ErrNotExist) {
			t.Errorf("Expected newly uploaded meta file to exist, it doesn't")
		}
	})
}
