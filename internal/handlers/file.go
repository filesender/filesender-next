package handlers

import (
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"

	"codeberg.org/filesender/filesender-next/internal/models"
)

// FileUpload handles a new file uploaded
func FileUpload(fileMeta models.File, file multipart.File) error {
	// Create transfer folder for user if not exists
	uploadDest := filepath.Join(os.Getenv("STATE_DIRECTORY"), fileMeta.UserID)
	if _, err := os.Stat(uploadDest); os.IsNotExist(err) {
		err = os.Mkdir(uploadDest, 0o700)
		if err != nil {
			return err
		}
	}

	dst, err := os.OpenFile(path.Join(uploadDest, fileMeta.ID+".tar"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
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

	err = fileMeta.Create()
	if err != nil {
		slog.Error("Failed creating upload meta file", "error", err)
		return err
	}

	return nil
}
