// Package handlers contains everything that has anything to do with handling (something)
// Handlers for API requests, template requests
// Handlers for file(s)
// Handlers for response
// Handlers for request
package handlers

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"

	"codeberg.org/filesender/filesender-next/internal/models"
)

// DownloadAPI handles `/api/v1/download/{userID}/{fileID}`
func DownloadAPI(stateDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, fileID := r.PathValue("userID"), r.PathValue("fileID")

		err := models.ValidateFile(stateDir, userID, fileID)
		if err != nil {
			slog.Error("User passed invalid file ID", "user_id", userID, "file_id", fileID, "error", err)
			sendError(w, http.StatusBadRequest, "File ID is invalid")
			return
		}

		file, err := models.GetFileFromIDs(stateDir, userID, fileID)
		if err != nil {
			slog.Error("Failed getting file from id", "user_id", userID, "file_id", fileID, "error", err)
			sendError(w, http.StatusInternalServerError, "Failed getting specified file")
			return
		}

		var offset int64
		offsetStr := r.Header.Get("Offset")
		if offsetStr != "" {
			offset, err = strconv.ParseInt(offsetStr, 10, 0)

			if err != nil {
				slog.Error("Failed parsing offset header to int", "offset", offsetStr, "error", err)
				sendJSON(w, http.StatusBadRequest, false, "Invalid Offset", nil)
				return
			}
		}

		if offset > file.ByteSize {
			slog.Error("Offset out of bounds!", "offset", offset, "file size", file.ByteSize)
			sendJSON(w, http.StatusBadRequest, false, "Offset is out of bounds", nil)
			return
		}

		filePath := filepath.Join(stateDir, userID, fileID+".bin")
		sendFileFromOffset(w, filePath, fileID+".bin", file.ByteSize, offset)
	}
}

// DownloadInfo handles `HEAD /api/v1/download/{userID}/{fileID}`
func DownloadInfo(stateDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, fileID := r.PathValue("userID"), r.PathValue("fileID")

		err := models.ValidateFile(stateDir, userID, fileID)
		if err != nil {
			slog.Error("User passed invalid file ID", "error", err)
			sendError(w, http.StatusBadRequest, "File ID is invalid")
			return
		}

		file, err := models.GetFileFromIDs(stateDir, userID, fileID)
		if err != nil {
			slog.Error("Failed getting file from id", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed getting specified file")
			return
		}

		w.Header().Add("Available", strconv.FormatBool(!file.Partial))
		w.Header().Add("File-Name", file.EncryptedFileName)

		w.Header().Add("Chunked", strconv.FormatBool(file.Chunked))
		w.Header().Add("Chunk-Count", strconv.FormatInt(int64(len(file.Chunks)), 10))
		w.Header().Add("Chunk-Size", strconv.FormatInt(file.ChunkSize, 10))
		w.Header().Add("Byte-Size", strconv.FormatInt(file.ByteSize, 10))

		sendEmptyResponse(w, http.StatusOK)
	}
}

// GetDownloadTemplate handles GET /download/{userID}/{fileID}
func GetDownloadTemplate(stateDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, fileID := r.PathValue("userID"), r.PathValue("fileID")

		err := models.ValidateFile(stateDir, userID, fileID)
		if err != nil {
			slog.Error("User passed invalid file ID", "error", err)
			sendError(w, http.StatusBadRequest, "File ID is invalid")
			return
		}

		file, err := models.GetFileFromIDs(stateDir, userID, fileID)
		if err != nil {
			slog.Error("Failed getting file from id", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed getting specified file")
			return
		}

		data := downloadTemplate{
			ByteSize: file.ByteSize,
			UserID:   userID,
			FileID:   fileID,
		}

		sendTemplate(w, "download", data)
	}
}
