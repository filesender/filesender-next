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

	"codeberg.org/filesender/filesender-next/internal/models"
)

// DownloadAPI handles `GET /api/download/{userID}/{fileID}`
func DownloadAPI(stateDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, fileID := r.PathValue("userID"), r.PathValue("fileID")

		err := models.ValidateFile(stateDir, userID, fileID)
		if err != nil {
			slog.Error("User passed invalid file ID", "user_id", userID, "file_id", fileID, "error", err)
			sendError(w, http.StatusBadRequest, "File ID is invalid")
			return
		}

		filePath := filepath.Join(stateDir, userID, fileID+".bin")
		http.ServeFile(w, r, filePath)
	}
}

// DownloadInfo handles `HEAD /api/download/{userID}/{fileID}`
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

		w.Header().Add("File-Name", file.EncryptedFileName)
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

		filePath := filepath.Join(stateDir, userID, fileID+".bin")
		byteSize, err := getFileSize(filePath)
		if err != nil {
			slog.Error("Failed getting file size", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed getting specified file")
			return
		}

		data := downloadTemplate{
			ByteSize: byteSize,
			UserID:   userID,
			FileID:   fileID,
		}

		sendTemplate(w, "download", data)
	}
}
