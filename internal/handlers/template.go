package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/id"
	"codeberg.org/filesender/filesender-next/internal/models"
	"codeberg.org/filesender/filesender-next/internal/utils"
)

// UploadTemplateHandler handles GET /{$}
func UploadTemplateHandler() http.HandlerFunc {
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

// UploadDoneTemplateHandler handles GET /upload/{id}
func UploadDoneTemplateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.Auth(r)
		if err != nil {
			slog.Info("unable to authenticate user", "error", err)
			sendError(w, http.StatusUnauthorized, "You're not authenticated")
			return
		}
		slog.Info("user authenticated", "user_id", userID)

		userID, err = utils.HashToBase64(userID)
		if err != nil {
			slog.Info("failed hashing user ID", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed creating user ID")
			return
		}

		transferID := r.PathValue("id")
		err = id.Validate(transferID)
		if err != nil {
			slog.Error("User passed invalid transfer ID", "error", err)
			sendError(w, http.StatusBadRequest, "Transfer ID is invalid")
			return
		}

		transfer, err := models.GetTransferFromIDs(userID, transferID)
		if err != nil {
			slog.Error("Failed getting transfer from id", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed getting specified transfer")
			return
		}

		if transfer.UserID != userID {
			sendError(w, http.StatusUnauthorized, "You're not authorized")
			return
		}

		sendTemplate(w, "upload_done", uploadDoneTemplate{
			UserID:     userID,
			TransferID: transfer.ID,
			FileCount:  transfer.FileCount,
			BytesSize:  transfer.TotalByteSize,
		})
	}
}

// GetTransferTemplateHandler handles GET /transfer/{userID}/{transferID}
func GetTransferTemplateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("userID")
		transferID := r.PathValue("transferID")
		err := id.Validate(transferID)
		if err != nil {
			slog.Error("User passed invalid transfer ID", "error", err)
			sendError(w, http.StatusBadRequest, "Transfer ID is invalid")
			return
		}

		err = models.TransferExists(userID, transferID)
		if err != nil {
			slog.Error("Could not find transfer", "error", err)
			sendError(w, http.StatusNotFound, "Could not find transfer")
			return
		}

		transfer, err := models.GetTransferFromIDs(userID, transferID)
		if err != nil {
			slog.Error("Failed getting transfer from id", "error", err)
			sendError(w, http.StatusInternalServerError, "Failed getting specified transfer")
			return
		}

		data := getTransferTemplate{
			FileCount: transfer.FileCount,
			ByteSize:  transfer.TotalByteSize,
		}

		for i := range transfer.FileCount {
			fileName := transfer.FileNames[i]
			file, err := models.GetFileFromName(userID, transferID, fileName)
			if err != nil {
				slog.Error("Failed getting file metadata", "error", err)
				continue
			}

			data.Files = append(data.Files, getTransferTemplateFile{
				FileName: fileName,
				ByteSize: file.ByteSize,
			})
		}

		sendTemplate(w, "transfer", data)
	}
}
