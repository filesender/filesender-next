// Package handlers contains everything that has anything to do with handling (something)
// Handlers for API requests, template requests
// Handlers for file(s)
// Handlers for response
// Handlers for request
package handlers

import (
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
)

// FileUpload handles a new file uploaded
func FileUpload(stateDir string, userID string, fileID string, file multipart.File) error {
	// Create transfer folder for user if not exists
	uploadDest := filepath.Join(stateDir, userID)
	if _, err := os.Stat(uploadDest); os.IsNotExist(err) {
		err = os.Mkdir(uploadDest, 0o700)
		if err != nil {
			slog.Error("Could not create new user directory", "error", err)
			return err
		}
	}

	fileName := fileID
	dst, err := os.OpenFile(path.Join(uploadDest, fileName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		if err := dst.Close(); err != nil {
			slog.Error("Failed closing file", "error", err)
		}
	}()

	_, err = io.Copy(dst, file)
	if err != nil {
		slog.Error("Failed copying file contents", "error", err)
		return err
	}

	return nil
}

// PartialFileUpload handles a chunk being uploaded
func PartialFileUpload(stateDir string, userID string, fileID string, file multipart.File, offset int64) (int64, error) {
	uploadDir := filepath.Join(stateDir, userID)
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		slog.Error("User upload directory does not exist", "path", uploadDir)
		return 0, err
	}

	filePath := filepath.Join(uploadDir, fileID)
	dst, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		slog.Error("Failed opening destination file", "error", err)
		return 0, err
	}
	defer func() {
		if err := dst.Close(); err != nil {
			slog.Error("Failed closing file", "error", err)
		}
	}()

	_, err = dst.Seek(offset, io.SeekStart)
	if err != nil {
		slog.Error("Failed seeking to offset", "offset", offset, "error", err)
		return 0, err
	}

	_, err = io.Copy(dst, file)
	if err != nil {
		slog.Error("Failed copying chunk data", "error", err)
		return 0, err
	}

	return getFileSize(filePath)
}

func getFileSize(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		slog.Error("Failed to get file info", "error", err)
		return 0, err
	}

	return fileInfo.Size(), nil
}
