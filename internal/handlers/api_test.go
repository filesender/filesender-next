package handlers_test

import (
	"bytes"
	"database/sql"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"codeberg.org/filesender/filesender-next/internal/config"
	"codeberg.org/filesender/filesender-next/internal/handlers"
	"codeberg.org/filesender/filesender-next/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

func TestCreateTransferAPIHandler(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:") // Using an in-memory database for testing
	defer db.Close()
	err := config.InitDB(db)
	if err != nil {
		t.Errorf("Failed initialising database: %v", err)
	}

	handler := handlers.CreateTransferAPIHandler(db)

	t.Run("Unauthenticated User", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/create", nil)
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
		}
	})

	t.Run("Valid Transfer Creation", func(t *testing.T) {
		requestBody := `{"subject":"Test Transfer","message":"Test Message","expiry_date":null}`
		req, _ := http.NewRequest("POST", "/create", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: "dummy_session"})
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.Code)
		}
	})
}

func TestUploadAPIHandler(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:") // Using an in-memory database for testing
	defer db.Close()
	err := config.InitDB(db)
	if err != nil {
		t.Errorf("Failed initialising database: %v", err)
	}

	handler := handlers.UploadAPIHandler(db, 10*1024*1024) // 10 MB limit

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
		writer.Close()

		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.AddCookie(&http.Cookie{Name: "session", Value: "dummy_session"})
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
		}
	})

	t.Run("Upload with valid transfer ID but no file", func(t *testing.T) {
		transfer := models.Transfer{
			UserID: "dummy_session",
		}
		err := transfer.Create(db)
		if err != nil {
			t.Fatalf("Failed creating new transfer: %v", err)
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("transfer_id", strconv.Itoa(1))
		writer.Close()

		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.AddCookie(&http.Cookie{Name: "session", Value: "dummy_session"})
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.Code)
		}
	})

	t.Run("Successful file upload", func(t *testing.T) {
		// Set a temporary directory
		tempDir, err := os.MkdirTemp("", "test_uploads")
		if err != nil {
			t.Fatalf("Failed to create temporary directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		os.Setenv("STATE_DIRECTORY", tempDir)

		transfer := models.Transfer{
			UserID: "dummy_session",
		}
		err = transfer.Create(db)
		if err != nil {
			t.Fatalf("Failed creating new transfer: %v", err)
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("transfer_id", strconv.Itoa(1))
		part, _ := writer.CreateFormFile("file", "testfile.txt")
		_, err = part.Write([]byte("This is a test file."))
		if err != nil {
			t.Errorf("Failed creating file contents")
		}
		writer.Close()

		req, _ := http.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.AddCookie(&http.Cookie{Name: "session", Value: "dummy_session"})
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		if resp.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.Code)
		}

		if _, err := os.Stat(path.Join(tempDir, "uploads", "1", "testfile.txt")); errors.Is(err, os.ErrNotExist) {
			t.Errorf("Expected newly uploaded file to exist, it doesn't")
		}

		if _, err := os.Stat(path.Join(tempDir, "uploads", "1", "testfile.txt.meta")); errors.Is(err, os.ErrNotExist) {
			t.Errorf("Expected newly uploaded meta file to exist, it doesn't")
		}
	})
}
