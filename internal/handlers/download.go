// Package handlers contains everything that has anything to do with handling (something)
// Handlers for API requests, template requests
// Handlers for file(s)
// Handlers for response
// Handlers for request
package handlers

import (
	"log/slog"
	"net/http"

	"codeberg.org/filesender/filesender-next/internal/models"
)

// DownloadAPI handles `GET /api/v1/download/{userID}/{fileID}`
func DownloadAPI(stateDir string) http.HandlerFunc {
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

		file.DownloadCount++
		err = file.Save(stateDir)
		if err != nil {
			slog.Error("Failed increasing download count on file", "error", err, "userID", userID, "fileID", fileID)
			sendError(w, http.StatusInternalServerError, "Failed setting new file meta data")
			return
		}

		sendFile(stateDir, w, &file)
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
			UserID:   file.UserID,
			FileID:   file.ID,
		}

		sendTemplate(w, "download", data)
	}
}
