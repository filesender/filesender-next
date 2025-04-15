package handlers

import (
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"codeberg.org/filesender/filesender-next/internal/models"
)

func getFullExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) <= 1 {
		return "" // no extension
	}
	return strings.Join(parts[1:], ".")
}

// FileUpload handles a new file uploaded
func FileUpload(stateDir string, fileMeta models.File, file multipart.File, fileName string) error {
	// Create transfer folder for user if not exists
	uploadDest := filepath.Join(stateDir, fileMeta.UserID)
	if _, err := os.Stat(uploadDest); os.IsNotExist(err) {
		err = os.Mkdir(uploadDest, 0o700)
		if err != nil {
			slog.Error("Could not create new user directory", "error", err)
			return err
		}
	}

	fileExtension := getFullExtension(fileName)
	fileName = fileMeta.ID + "." + fileExtension
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

	fileMeta.FileName = fileName
	err = fileMeta.Create(stateDir)
	if err != nil {
		slog.Error("Failed creating upload meta file", "error", err)
		return err
	}

	return nil
}

// PartialFileUpload handles a chunk being uploaded
func PartialFileUpload(stateDir string, fileMeta models.File, file multipart.File, offset int64) error {
	// This should already exist
	uploadDest := filepath.Join(stateDir, fileMeta.UserID)

	// But more specifically, this could maybe not exist already...
	uploadDest = filepath.Join(uploadDest, fileMeta.ID)
	if _, err := os.Stat(uploadDest); os.IsNotExist(err) {
		err = os.Mkdir(uploadDest, 0o700)
		if err != nil {
			slog.Error("Could not create new chunked directory", "error", err)
			return err
		}
	}

	hexOffset := strconv.FormatInt(offset, 16)
	dst, err := os.OpenFile(path.Join(uploadDest, hexOffset+".bin"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
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

	// TODO: Maybe do something with meta to keep track of which offsets & sizes exist?
	// We could also just read directory contents of course..

	return nil
}
