package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"codeberg.org/filesender/filesender-next/internal/assets"
	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/handlers"
	"codeberg.org/filesender/filesender-next/internal/hash"
)

func mockRequest(handler http.HandlerFunc, method string, url string, headers map[string]string, pathValues map[string]string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, url, nil)
	resp := httptest.NewRecorder()

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	for k, v := range pathValues {
		req.SetPathValue(k, v)
	}

	handler.ServeHTTP(resp, req)
	return resp
}

func TestGetDownloadTemplate(t *testing.T) {
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
	err = hash.Init(tempDir)
	if err != nil {
		t.Fatalf("Could not initialise hashing package: %v", err)
	}

	body, writer := createMultipartBody("Hello, world!")
	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed closing writer: %v", err)
	}

	resp := mockUploadRequest(handler, body, writer, nil)
	loc := resp.Header().Get("Location")

	handler = handlers.GetDownloadTemplate("/", tempDir)
	locSplits := strings.Split(loc, "/")
	userID, fileID := locSplits[2], locSplits[3]
	println(userID, fileID)

	t.Run("User not exist", func(t *testing.T) {
		resp := mockRequest(handler, "GET", fmt.Sprintf("/view/hi/%s", fileID), nil, map[string]string{
			"userID": "hi",
			"fileID": fileID,
		})

		if resp.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "Failed getting specified file") {
			t.Errorf("Expected error to be \"Failed getting specified file\", got %s", b)
		}
	})

	t.Run("File not exist", func(t *testing.T) {
		resp := mockRequest(handler, "GET", fmt.Sprintf("/view/%s/hi", fileID), nil, map[string]string{
			"userID": userID,
			"fileID": "hi",
		})

		if resp.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "Failed getting specified file") {
			t.Errorf("Expected error to be \"Failed getting specified file\", got %s", b)
		}
	})

	t.Run("File & user not exist", func(t *testing.T) {
		resp := mockRequest(handler, "GET", "/view/hi/hi", nil, map[string]string{
			"userID": "hi",
			"fileID": "hi",
		})

		if resp.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "Failed getting specified file") {
			t.Errorf("Expected error to be \"Failed getting specified file\", got %s", b)
		}
	})

	t.Run("Template error", func(t *testing.T) {
		resp := mockRequest(handler, "GET", fmt.Sprintf("/view/%s/%s", userID, fileID), nil, map[string]string{
			"userID": userID,
			"fileID": fileID,
		})

		if resp.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "Error rendering page") {
			t.Errorf("Expected error to be \"Error rendering page\", got \"%s\"", b)
		}
	})

	t.Run("Success", func(t *testing.T) {
		handlers.Init(assets.EmbeddedTemplateFiles)

		resp := mockRequest(handler, "GET", fmt.Sprintf("/view/%s/%s", userID, fileID), nil, map[string]string{
			"userID": userID,
			"fileID": fileID,
		})

		if resp.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "1 file (13 bytes)") {
			t.Errorf("Expected response to contain \"1 file (13 bytes)\", got \"%s\"", b)
		}
	})
}

func TestUploadTemplate(t *testing.T) {
	t.Run("Not authenticated", func(t *testing.T) {
		handler := handlers.UploadTemplate("/", &auth.ProxyAuth{})
		resp := mockRequest(handler, "GET", "/", nil, nil)

		if resp.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "You're not authenticated") {
			t.Errorf("Expected error to be \"You're not authenticated\", got \"%s\"", b)
		}
	})

	t.Run("Success", func(t *testing.T) {
		handler := handlers.UploadTemplate("/", &auth.DummyAuth{})
		resp := mockRequest(handler, "GET", "/", nil, nil)

		if resp.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, resp.Code)
		}
	})
}
