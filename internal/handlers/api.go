package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/models"
)

// Creates a transfer, returns a transfer object
// This should be called before uploading
func CreateTransferAPIHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.Auth(r)
		if err != nil {
			slog.Info("unable to authenticate user", "error", err)
			sendJSON(w, http.StatusUnauthorized, false, "You're not authenticated", nil)
			return
		}
		slog.Info("user authenticated", "user_id", userID)

		// Read requets body
		var requestBody createTransferAPIRequest
		success := recvJSON(w, r, &requestBody)
		if !success {
			return
		}

		// Handle nil pointer
		now := time.Now()
		expiryDate := now.Add(time.Hour * 24 * 7)
		if requestBody.ExpiryDate != nil {
			expiryDate = *requestBody.ExpiryDate
		}

		if expiryDate.Before(now) || expiryDate.After(now.AddDate(0, 0, 30)) {
			sendJSON(w, http.StatusBadRequest, false, "Expiry date must be in the future, but max 30 days in the future", nil)
			return
		}

		transfer := models.Transfer{
			UserID:     userID,
			Subject:    requestBody.Subject,
			Message:    requestBody.Message,
			ExpiryDate: expiryDate,
		}
		err = transfer.Create(db)
		if err != nil {
			slog.Error("Failed creating transfer", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed creating transfer", nil)
			return
		}

		slog.Debug("Successfully created new transfer", "user", userID, "transfer", transfer.ID)
		sendJSON(w, http.StatusCreated, true, "", createTransferAPIResponse{
			Transfer: transfer,
		})
	}
}

// Handles file upload to specific transfer
func UploadAPIHandler(db *sql.DB, maxUploadSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.Auth(r)
		if err != nil {
			slog.Info("unable to authenticate user", "error", err)
			sendJSON(w, http.StatusUnauthorized, false, "You're not authenticated", nil)
			return
		}
		slog.Info("user authenticated", "user_id", userID)

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			sendJSON(w, http.StatusRequestEntityTooLarge, false, "Upload file size too large", nil) // there's no "payload too large" in http std
			return
		}

		transferIDs := r.MultipartForm.Value["transfer_id"]
		if len(transferIDs) != 1 {
			sendJSON(w, http.StatusBadRequest, false, "Expected a transfer id", nil)
			return
		}

		transferID, err := strconv.ParseInt(transferIDs[0], 10, 0)
		if err != nil {
			sendJSON(w, http.StatusBadRequest, false, "Transfer ID is not a number", nil)
			return
		}

		relativePath := ""
		relativePaths := r.MultipartForm.Value["relative_path"]
		if len(relativePaths) == 1 {
			relativePath = filepath.Clean(relativePaths[0])
			if strings.Contains(relativePath, "..") {
				slog.Error("Upload invalid relative path: trying to access parent directory")
				sendJSON(w, http.StatusBadRequest, false, "Invalid relative path", nil)
				return
			}
		}

		transfer, err := models.GetTransferFromID(db, int(transferID))
		if err != nil {
			sendJSON(w, http.StatusNotFound, false, "Could not find the transfer", nil)
			return
		}
		if transfer.UserID != userID {
			sendJSON(w, http.StatusUnauthorized, false, "You're not authorized to modify this transfer", nil)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			slog.Error("Failed opening file", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Lost the file", nil)
			return
		}
		defer file.Close()

		err = transfer.NewFile(db, int(fileHeader.Size))
		if err != nil {
			slog.Error("Failed adding new file to transfer object", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Handle file failed", nil)
			return
		}

		err = HandleFileUpload(transfer, file, fileHeader, relativePath)
		if err != nil {
			slog.Error("Failed handling newly uploaded file!", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Handle file failed", nil)

			// Remove the file from transfer object if handling file failed
			err = transfer.RemoveFile(db, int(fileHeader.Size))
			if err != nil {
				slog.Error("Failed removing file from transfer object", "error", err)
			}
			return
		}

		slog.Debug("Successfully created new file", "user", userID, "transfer", transfer.ID, "file", fileHeader.Filename)
		sendJSON(w, http.StatusCreated, true, "", nil)
	}
}
