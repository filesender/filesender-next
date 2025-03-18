package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"codeberg.org/filesender/filesender-next/internal/models"
)

// Handles newly uploaded file
func HandleFileUpload(transfer models.Transfer, file multipart.File, fileHeader *multipart.FileHeader, relativePath string) error {
	// Create uploads folder if not exist
	uploadsPath := path.Join(os.Getenv("STATE_DIRECTORY"), "uploads")
	if _, err := os.Stat(uploadsPath); os.IsNotExist(err) {
		err = os.Mkdir(uploadsPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Create transfer folder if not exists
	baseUploadDir := path.Join(uploadsPath, strconv.Itoa(transfer.ID))
	if _, err := os.Stat(baseUploadDir); os.IsNotExist(err) {
		err = os.Mkdir(baseUploadDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	uploadDest := ""
	if relativePath == "" { // If relative path not set, default to base dir
		uploadDest = baseUploadDir

	} else { // Else check if relative path is in base dir & create if not exists
		uploadDest = filepath.Join(baseUploadDir, relativePath)
		if !strings.HasPrefix(uploadDest, baseUploadDir) {
			return fmt.Errorf("upload invalid relative path: trying to access outside of upload destination")
		}

		if _, err := os.Stat(uploadDest); os.IsNotExist(err) {
			err = os.MkdirAll(uploadDest, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	dst, err := os.Create(path.Join(uploadDest, fileHeader.Filename))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return err
	}

	meta := models.File{
		DownloadCount: 0,
	}
	err = meta.Create(uploadDest, fileHeader.Filename)
	if err != nil {
		return err
	}

	return nil
}
