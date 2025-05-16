package handlers

import (
	"log/slog"
	"net/http"
	"path/filepath"

	"codeberg.org/filesender/filesender-next/internal/auth"
)

// GetDownloadTemplate handles GET /view/{userID}/{fileID}
func GetDownloadTemplate(appRoot string, stateDir string) http.HandlerFunc {
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
			AppRoot:  appRoot,
			ByteSize: byteSize,
			UserID:   userID,
			FileID:   fileID,
		}

		sendTemplate(w, "download", data)
	}
}

// UploadTemplate handles GET /{$}
func UploadTemplate(appRoot string, authModule auth.Auth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authModule.UserAuth(r)
		if err != nil {
			slog.Info("unable to authenticate user", "error", err)
			sendError(w, http.StatusUnauthorized, "You're not authenticated")
			return
		}
		slog.Info("user authenticated", "user_id", userID)

		sendTemplate(w, "upload", uploadTemplate{
			AppRoot: appRoot,
		})
	}
}
