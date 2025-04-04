package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/crypto"
	"codeberg.org/filesender/filesender-next/internal/id"
	"codeberg.org/filesender/filesender-next/internal/models"
)

// UploadTemplate handles GET /{$}
func UploadTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.Auth(r)
		if err != nil {
			slog.Info("unable to authenticate user", "error", err)
			sendError(w, http.StatusUnauthorized, "You're not authenticated")
			return
		}
		slog.Info("user authenticated", "user_id", userID)

		minDate := time.Now().UTC().Add(time.Hour * 24)
		defaultDate := time.Now().UTC().Add(time.Hour * 24 * 7)
		maxDate := time.Now().UTC().Add(time.Hour * 24 * 30)

		sendTemplate(w, "upload", uploadTemplate{
			MinDate:     minDate.Format(time.DateOnly),
			DefaultDate: defaultDate.Format(time.DateOnly),
			MaxDate:     maxDate.Format(time.DateOnly),
		})
	}
}

// UploadDoneTemplate handles GET /upload/{id}
func UploadDoneTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.Auth(r)
		if err != nil {
			slog.Info("unable to authenticate user", "error", err)
			sendError(w, http.StatusUnauthorized, "You're not authenticated")
			return
		}
		slog.Info("user authenticated", "user_id", userID)

		userID, err = crypto.HashToBase64(userID)
		if err != nil {
			slog.Info("failed hashing user ID", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed creating user ID")
			return
		}

		fileID := r.PathValue("id")
		err = id.Validate(fileID)
		if err != nil {
			slog.Error("User passed invalid file ID", "error", err)
			sendError(w, http.StatusBadRequest, "File ID is invalid")
			return
		}

		file, err := models.GetFileFromIDs(userID, fileID)
		if err != nil {
			slog.Error("Failed getting file from id", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed getting specified file")
			return
		}

		if file.UserID != userID {
			sendError(w, http.StatusUnauthorized, "You're not authorized")
			return
		}

		sendTemplate(w, "upload_done", uploadDoneTemplate{
			UserID:    file.UserID,
			FileID:    file.ID,
			BytesSize: file.ByteSize,
		})
	}
}

// GetDownloadTemplate handles GET /download/{userID}/{fileID}
func GetDownloadTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, fileID := r.PathValue("userID"), r.PathValue("fileID")

		err := models.ValidateFile(userID, fileID)
		if err != nil {
			slog.Error("User passed invalid file ID", "error", err)
			sendError(w, http.StatusBadRequest, "File ID is invalid")
			return
		}

		file, err := models.GetFileFromIDs(userID, fileID)
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
