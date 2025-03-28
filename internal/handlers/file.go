package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"

	"codeberg.org/filesender/filesender-next/internal/models"
)

// FileUpload handles a new file uploaded
func FileUpload(transfer models.Transfer, file multipart.File, fileHeader *multipart.FileHeader) error {
	// Create uploads folder if not exist
	uploadsPath := path.Join(os.Getenv("STATE_DIRECTORY"), "uploads")
	if _, err := os.Stat(uploadsPath); os.IsNotExist(err) {
		err = os.Mkdir(uploadsPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Create transfer folder if not exists
	uploadDest := filepath.Clean(path.Join(uploadsPath, transfer.UserID, transfer.ID))
	if _, err := os.Stat(uploadDest); os.IsNotExist(err) {
		err = os.Mkdir(uploadDest, os.ModePerm)
		if err != nil {
			return err
		}
	}

	dst, err := os.Create(path.Join(uploadDest, fileHeader.Filename))
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
		return err
	}

	meta := models.File{
		ByteSize: int(fileHeader.Size),
	}
	err = meta.Create(transfer.UserID, transfer.ID, fileHeader.Filename)
	if err != nil {
		return err
	}

	return nil
}

func getFile(transfer models.Transfer, fileName string) (models.File, error) {
	var file models.File

	exists, err := models.FileExists(transfer.UserID, transfer.ID, fileName)
	if err != nil {
		return file, err
	}

	if !exists {
		slog.Error("File does not exist!", "userID", transfer.UserID, "transferID", transfer.ID, "fileName", fileName)
		return file, fmt.Errorf("file does not exist: %s", fileName)
	}

	file, err = models.GetFileFromName(transfer.UserID, transfer.ID, fileName)
	if err != nil {
		return file, err
	}

	return file, nil
}
