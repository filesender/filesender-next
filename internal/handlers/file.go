package handlers

import (
	"io"
	"mime/multipart"
	"os"
	"path"
	"strconv"

	"codeberg.org/filesender/filesender-next/internal/models"
)

// Handles newly uploaded file
func HandleFileUpload(transfer models.Transfer, file multipart.File, fileHeader *multipart.FileHeader) error {
	err := os.Mkdir(path.Join(os.Getenv("STATE_DIRECTORY"), "uploads"), os.ModePerm)
	if err != nil {
		return err
	}

	uploadDest := path.Join(os.Getenv("STATE_DIRECTORY"), "uploads", strconv.Itoa(transfer.ID))
	err = os.Mkdir(uploadDest, os.ModePerm)
	if err != nil {
		return err
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
