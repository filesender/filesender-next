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
)

// GetDownloadTemplate handles GET /view/{userID}/{fileID}
func GetDownloadTemplate(stateDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, fileID := r.PathValue("userID"), r.PathValue("fileID")

		filePath := filepath.Join(stateDir, userID, fileID)
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
