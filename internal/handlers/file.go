package handlers

import (
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"

	"codeberg.org/filesender/filesender-next/internal/models"
)

// FileUpload handles a new file uploaded
func FileUpload(stateDir string, userID string, fileID string, fileMeta models.File, file multipart.File) error {
	// Create transfer folder for user if not exists
	uploadDest := filepath.Join(stateDir, userID)
	if _, err := os.Stat(uploadDest); os.IsNotExist(err) {
		err = os.Mkdir(uploadDest, 0o700)
		if err != nil {
			slog.Error("Could not create new user directory", "error", err)
			return err
		}
	}

	fileName := fileID + ".bin"
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

	err = fileMeta.Create(stateDir, userID, fileID)
	if err != nil {
		slog.Error("Failed creating upload meta file", "error", err)
		return err
	}

	return nil
}

// PartialFileUpload handles a chunk being uploaded
func PartialFileUpload(stateDir string, userID string, fileID string, fileMeta *models.File, file multipart.File, offset int64) error {
	// This should already exist
	uploadDest := filepath.Join(stateDir, userID)

	// But more specifically, this could maybe not exist already...
	uploadDest = filepath.Join(uploadDest, fileID)
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
	fileMeta.Chunks = append(fileMeta.Chunks, hexOffset)

	// Make sure the chunks are sorted in the array
	sort.Slice(fileMeta.Chunks, func(i int, j int) bool {
		a, _ := strconv.ParseInt(fileMeta.Chunks[i], 16, 64)
		b, _ := strconv.ParseInt(fileMeta.Chunks[j], 16, 64)
		return a < b
	})

	return nil
}
