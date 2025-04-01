package handlers

import (
	"archive/zip"
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
	// Create transfer folder if not exists
	uploadDest := filepath.Clean(path.Join(os.Getenv("STATE_DIRECTORY"), transfer.UserID, transfer.ID))
	if _, err := os.Stat(uploadDest); os.IsNotExist(err) {
		err = os.Mkdir(uploadDest, 0o700)
		if err != nil {
			return err
		}
	}

	dst, err := os.OpenFile(path.Join(uploadDest, fileHeader.Filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
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
	err = meta.Save(transfer.UserID, transfer.ID, fileHeader.Filename)
	if err != nil {
		return err
	}

	return nil
}

func addFileToZip(zipWriter *zip.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil {
		slog.Error("Failed opening file", "file", path, "error", err)
		return err
	}
	defer func() {
		err := f.Close()
		if err != nil {
			slog.Error("Failed closing file", "file", path, "error", err)
		}
	}()

	wr, err := zipWriter.Create(filepath.Base(path))
	if err != nil {
		slog.Error("Failed to create zip entry", "file", path, "error", err)
		return err
	}

	_, err = io.Copy(wr, f)
	if err != nil {
		slog.Error("Failed to write file to zip", "file", path, "error", err)
	}
	return err
}

func getFile(transfer *models.Transfer, fileName string) (models.File, error) {
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
