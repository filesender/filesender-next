package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/models"
	"github.com/google/uuid"
)

func UploadTemplateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		minDate := time.Now().UTC().Add(time.Hour * 24)
		defaultDate := time.Now().UTC().Add(time.Hour * 24 * 7)
		maxDate := time.Now().UTC().Add(time.Hour * 24 * 30)

		data := map[string]any{
			"MinDate":     minDate.Format(time.DateOnly),
			"DefaultDate": defaultDate.Format(time.DateOnly),
			"MaxDate":     maxDate.Format(time.DateOnly),
		}

		sendTemplate(w, "upload", data)
	}
}

func UploadDoneTemplateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.Auth(r)
		if err != nil {
			slog.Info("unable to authenticate user", "error", err)
			sendError(w, http.StatusUnauthorized, "You're not authenticated")
			return
		}
		slog.Info("user authenticated", "user_id", userID)

		transferID := r.PathValue("id")
		err = uuid.Validate(transferID)
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

		data := map[string]any{
			"TransferID": transfer.ID,
			"FileCount":  transfer.FileCount,
			"BytesSize":  transfer.TotalByteSize,
		}

		sendTemplate(w, "upload_done", data)
	}
}
