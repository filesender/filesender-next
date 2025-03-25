// Package handlers contains everything that has anything to do with handling (something)
// Handlers for API requests, template requests
// Handlers for file(s)
// Handlers for response
// Handlers for request
package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"codeberg.org/filesender/filesender-next/internal/auth"
	"codeberg.org/filesender/filesender-next/internal/id"
	"codeberg.org/filesender/filesender-next/internal/models"
)

// CreateTransferAPIHandler handles POST /api/v1/transfers
// Creates a transfer, returns a transfer object
// This should be called before uploading
func CreateTransferAPIHandler() http.HandlerFunc {
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
			ExpiryDate: expiryDate,
		}
		err = transfer.Create()
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

// UploadAPIHandler handles POST /api/v1/upload
func UploadAPIHandler(maxUploadSize int64) http.HandlerFunc {
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
		transferID := transferIDs[0]
		err = id.Validate(transferID)
		if err != nil {
			sendJSON(w, http.StatusBadRequest, false, "Transfer ID is incorrectly formatted", nil)
			return
		}

		transfer, err := models.GetTransferFromIDs(userID, transferID)
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
		defer func() {
			if err := file.Close(); err != nil {
				slog.Error("Failed closing file", "error", err)
			}
		}()

		err = HandleFileUpload(transfer, file, fileHeader)
		if err != nil {
			slog.Error("Failed handling newly uploaded file!", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Handle file failed", nil)
			return
		}

		err = transfer.NewFile(int(fileHeader.Size))
		if err != nil {
			slog.Error("Failed adding new file to transfer object", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Handle file failed", nil)
			return
		}

		slog.Debug("Successfully created new file", "user", userID, "transfer", transfer.ID, "file", fileHeader.Filename)
		sendJSON(w, http.StatusCreated, true, "", nil)
	}
}
