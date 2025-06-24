package handlers_test

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/handlers"
	"codeberg.org/filesender/filesender-next/internal/hash"
)

func createMultipartBody(fileBody string) (*bytes.Buffer, *multipart.Writer) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "testfile.txt")
	_, _ = part.Write([]byte(fileBody))
	return body, writer
}

func mockUploadRequest(handler http.HandlerFunc, body *bytes.Buffer, writer *multipart.Writer, headers map[string]string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	handler.ServeHTTP(resp, req)
	return resp
}

func mockPartialUploadRequest(handler http.HandlerFunc, fileID string, body *bytes.Buffer, writer *multipart.Writer, headers map[string]string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/upload/%s", fileID), body)
	req.SetPathValue("fileID", fileID)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	handler.ServeHTTP(resp, req)
	return resp
}

func clearFolder(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		err := os.Remove(filepath.Join(dir, entry.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

func createFile(path string, body string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(body)
	return err
}

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
	err = hash.Init(tempDir)
	if err != nil {
		t.Fatalf("Could not initialise hashing package: %v", err)
	}

	// Hash "dev" for test use
	hashedID, err := hash.ToBase64("dev")
	if err != nil {
		t.Fatalf("Failed hashing dummy user ID: %v", err)
	}

	t.Run("Fail authentication", func(t *testing.T) {
		handler := handlers.UploadAPI("/", &auth.ProxyAuth{}, tempDir, 10*1024*1024)
		body, writer := createMultipartBody("")
		writer.Close()

		resp := mockUploadRequest(handler, body, writer, nil)
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "You're not authenticated") {
			t.Errorf("Expected error to be \"You're not authenticated\", got %s", b)
		}
	})

	t.Run("Fail hashing user ID", func(t *testing.T) {
		hash.ResetKeyForTest()
		defer hash.Init(tempDir)

		body, writer := createMultipartBody("")
		writer.Close()

		resp := mockUploadRequest(handler, body, writer, nil)
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "Failed creating user ID") {
			t.Errorf("Expected error to be \"Failed creating user ID\", got %s", b)
		}
	})

	t.Run("Too big file size", func(t *testing.T) {
		handler := handlers.UploadAPI("/", &auth.DummyAuth{}, tempDir, 10)
		body, writer := createMultipartBody("Hello, world!")
		writer.Close()

		resp := mockUploadRequest(handler, body, writer, nil)
		if resp.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("Expected status %d, got %d", http.StatusRequestEntityTooLarge, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "Upload file size too large") {
			t.Errorf("Expected error to be \"Upload file size too large\", got %s", b)
		}
	})

	t.Run("Upload with no file", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.Close()

		resp := mockUploadRequest(handler, body, writer, nil)
		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "No file") {
			t.Errorf("Expected error to be \"No file\", got %s", b)
		}
	})

	t.Run("Successful file upload", func(t *testing.T) {
		defer clearFolder(filepath.Join(tempDir, hashedID))

		body, writer := createMultipartBody("Hello, world!")
		writer.Close()

		resp := mockUploadRequest(handler, body, writer, nil)
		if resp.Code != http.StatusSeeOther {
			t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.Code)
		}

		files, err := os.ReadDir(path.Join(tempDir, hashedID))
		if err != nil {
			t.Fatalf("Failed to read directory: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("Expected 1 file in directory, found %d", len(files))
		}

		locationHeader := resp.Header().Get("Location")
		if locationHeader == "" {
			t.Errorf("There was no location header in response!")
		} else if !strings.Contains(locationHeader, hashedID) {
			t.Errorf("Expected location header to contain \"%s\", got \"%s\"", hashedID, locationHeader)
		} else if strings.Count(locationHeader, "/") != 3 {
			t.Errorf("Expected location header to contain 3 `/`, instead got %d: \"%s\"", strings.Count(locationHeader, "/"), locationHeader)
		} else if !strings.Contains(locationHeader, "/view/") {
			t.Errorf("Expected location header to contain \"/view/\", instead fot \"%s\"", locationHeader)
		}
	})

	t.Run("Successful partial file upload", func(t *testing.T) {
		defer clearFolder(filepath.Join(tempDir, hashedID))

		body, writer := createMultipartBody("Hello, world!")
		writer.Close()

		resp := mockUploadRequest(handler, body, writer, map[string]string{
			"Upload-Complete": "0",
		})
		if resp.Code != http.StatusAccepted {
			t.Errorf("Expected status %d, got %d", http.StatusAccepted, resp.Code)
		}

		files, err := os.ReadDir(path.Join(tempDir, hashedID))
		if err != nil {
			t.Fatalf("Failed to read directory: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("Expected 1 file in directory, found %d (%v)", len(files), files)
		}

		locationHeader := resp.Header().Get("Location")
		if locationHeader == "" {
			t.Errorf("There was no location header in response!")
		} else if !strings.Contains(locationHeader, "/upload/") {
			t.Errorf("Expected location header to contain \"/upload/\", got \"%s\"", locationHeader)
		} else if strings.Count(locationHeader, "/") != 2 {
			t.Errorf("Expected location header to contain 2 `/`, instead got %d: \"%s\"", strings.Count(locationHeader, "/"), locationHeader)
		}
	})
}

func TestChunkedUploadAPIHandler(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_uploads")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed closing file %v", err)
		}
	}()

	handler := handlers.ChunkedUploadAPI("/", &auth.DummyAuth{}, tempDir, 10*1024*1024) // 10 MB limit
	err = hash.Init(tempDir)
	if err != nil {
		t.Fatalf("Could not initialise hashing package: %v", err)
	}

	// Hash "dev" for test use
	hashedID, err := hash.ToBase64("dev")
	if err != nil {
		t.Fatalf("Failed hashing dummy user ID: %v", err)
	}

	t.Run("No file ID", func(t *testing.T) {
		body, writer := createMultipartBody("")
		writer.Close()

		resp := mockPartialUploadRequest(handler, "", body, writer, nil)
		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "File ID expected") {
			t.Errorf("Expected error to be \"File ID expected\", got \"%s\"", b)
		}
	})

	t.Run("Fail authentication", func(t *testing.T) {
		handler := handlers.ChunkedUploadAPI("/", &auth.ProxyAuth{}, tempDir, 10*1024*1024)
		body, writer := createMultipartBody("")
		writer.Close()

		resp := mockPartialUploadRequest(handler, "id", body, writer, nil)
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "You're not authenticated") {
			t.Errorf("Expected error to be \"You're not authenticated\", got \"%s\"", b)
		}
	})

	t.Run("Fail hashing user ID", func(t *testing.T) {
		hash.ResetKeyForTest()
		defer hash.Init(tempDir)

		body, writer := createMultipartBody("")
		writer.Close()

		resp := mockPartialUploadRequest(handler, "id", body, writer, nil)
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "Failed creating user ID") {
			t.Errorf("Expected error to be \"Failed creating user ID\", got %s", b)
		}
	})

	t.Run("Too big file size", func(t *testing.T) {
		handler := handlers.ChunkedUploadAPI("/", &auth.DummyAuth{}, tempDir, 10)
		body, writer := createMultipartBody("Hello, world!")
		writer.Close()

		resp := mockPartialUploadRequest(handler, "id", body, writer, nil)
		if resp.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("Expected status %d, got %d", http.StatusRequestEntityTooLarge, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "Upload file size too large") {
			t.Errorf("Expected error to be \"Upload file size too large\", got %s", b)
		}
	})

	t.Run("Missing upload offset", func(t *testing.T) {
		body, writer := createMultipartBody("Hello, world!")
		writer.Close()

		resp := mockPartialUploadRequest(handler, "id", body, writer, nil)
		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "Missing offset") {
			t.Errorf("Expected error to be \"Missing offset\", got %s", b)
		}
	})

	t.Run("Invalid upload offset", func(t *testing.T) {
		body, writer := createMultipartBody("Hello, world!")
		writer.Close()

		resp := mockPartialUploadRequest(handler, "id", body, writer, map[string]string{
			"Upload-Offset": "fail lol",
		})
		if resp.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "Invalid offset") {
			t.Errorf("Expected error to be \"Invalid offset\", got %s", b)
		}
	})

	t.Run("Upload with no file", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.Close()

		resp := mockPartialUploadRequest(handler, "id", body, writer, map[string]string{
			"Upload-Offset": "1",
		})
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "Lost the file") {
			t.Errorf("Expected error to be \"Lost the file\", got %s", b)
		}
	})

	t.Run("File does not exist", func(t *testing.T) {
		body, writer := createMultipartBody("Hello, world!")
		writer.Close()

		resp := mockPartialUploadRequest(handler, "id", body, writer, map[string]string{
			"Upload-Offset": "1",
		})
		if resp.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, resp.Code)
		}

		b := resp.Body.Bytes()
		if !strings.Contains(string(b), "Failed handling existing file upload") {
			t.Errorf("Expected error to be \"Failed handling existing file upload\", got %s", b)
		}
	})

	t.Run("Success partial upload", func(t *testing.T) {
		defer clearFolder(filepath.Join(tempDir, hashedID))

		err = createFile(filepath.Join(tempDir, hashedID, "file_id"), "Hello, ")
		body, writer := createMultipartBody("world!")
		writer.Close()

		resp := mockPartialUploadRequest(handler, "file_id", body, writer, map[string]string{
			"Upload-Offset": "7",
		})
		if resp.Code != http.StatusSeeOther {
			t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.Code)

			b := resp.Body.Bytes()
			t.Errorf("Response bytes: %s", b)
		}

		files, err := os.ReadDir(path.Join(tempDir, hashedID))
		if err != nil {
			t.Fatalf("Failed to read directory: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("Expected 1 file in directory, found %d", len(files))
		}

		locationHeader := resp.Header().Get("Location")
		if locationHeader == "" {
			t.Errorf("There was no location header in response!")
		} else if !strings.Contains(locationHeader, hashedID) {
			t.Errorf("Expected location header to contain \"%s\", got \"%s\"", hashedID, locationHeader)
		} else if strings.Count(locationHeader, "/") != 3 {
			t.Errorf("Expected location header to contain 3 `/`, instead got %d: \"%s\"", strings.Count(locationHeader, "/"), locationHeader)
		} else if !strings.Contains(locationHeader, "/view/") {
			t.Errorf("Expected location header to contain \"/view/\", instead fot \"%s\"", locationHeader)
		}

		b, err := os.ReadFile(filepath.Join(tempDir, hashedID, files[0].Name()))
		if err != nil {
			t.Fatalf("Failed opening file! %v", err)
		}

		if string(b) != "Hello, world!" {
			t.Errorf("Expected file contents to be \"Hello, world!\", got \"%s\"", b)
		}
	})

	t.Run("Success partial upload, non complete", func(t *testing.T) {
		defer clearFolder(filepath.Join(tempDir, hashedID))

		err = createFile(filepath.Join(tempDir, hashedID, "file_id"), "Hello, ")
		body, writer := createMultipartBody("world!")
		writer.Close()

		resp := mockPartialUploadRequest(handler, "file_id", body, writer, map[string]string{
			"Upload-Offset":   "7",
			"Upload-Complete": "0",
		})
		if resp.Code != http.StatusAccepted {
			t.Errorf("Expected status %d, got %d", http.StatusAccepted, resp.Code)

			b := resp.Body.Bytes()
			t.Errorf("Response bytes: %s", b)
		}

		files, err := os.ReadDir(path.Join(tempDir, hashedID))
		if err != nil {
			t.Fatalf("Failed to read directory: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("Expected 1 file in directory, found %d", len(files))
		}

		locationHeader := resp.Header().Get("Location")
		if locationHeader == "" {
			t.Errorf("There was no location header in response!")
		} else if !strings.Contains(locationHeader, "/upload/") {
			t.Errorf("Expected location header to contain \"/upload/\", got \"%s\"", locationHeader)
		} else if strings.Count(locationHeader, "/") != 2 {
			t.Errorf("Expected location header to contain 2 `/`, instead got %d: \"%s\"", strings.Count(locationHeader, "/"), locationHeader)
		}

		b, err := os.ReadFile(filepath.Join(tempDir, hashedID, files[0].Name()))
		if err != nil {
			t.Fatalf("Failed opening file! %v", err)
		}

		if string(b) != "Hello, world!" {
			t.Errorf("Expected file contents to be \"Hello, world!\", got \"%s\"", b)
		}
	})
}
