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
	"codeberg.org/filesender/filesender-next/internal/crypto"
	"codeberg.org/filesender/filesender-next/internal/id"
	"codeberg.org/filesender/filesender-next/internal/models"
)

// UploadAPI handles POST /api/v1/upload
// Expects either `transfer_id` or `expiry_date` in form data
// If `transfer_id` is not set, it creates a new transfer.
func UploadAPI(maxUploadSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.Auth(r)
		if err != nil {
			slog.Info("unable to authenticate user", "error", err)
			sendJSON(w, http.StatusUnauthorized, false, "You're not authenticated", nil)
			return
		}
		slog.Info("user authenticated", "user_id", userID)

		userID, err = crypto.HashToBase64(userID)
		if err != nil {
			slog.Info("failed hashing user ID", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed creating user ID", nil)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			sendJSON(w, http.StatusRequestEntityTooLarge, false, "Upload file size too large", nil)
			return
		}

		var transfer models.Transfer
		transferIDs := r.MultipartForm.Value["transfer-id"]
		if len(transferIDs) == 1 {
			transferID := transferIDs[0]
			err = id.Validate(transferID)
			if err != nil {
				sendJSON(w, http.StatusBadRequest, false, "Transfer ID is incorrectly formatted", nil)
				return
			}

			transfer, err = models.GetTransferFromIDs(userID, transferID)
			if err != nil {
				sendJSON(w, http.StatusNotFound, false, "Could not find the transfer", nil)
				return
			}
			if transfer.UserID != userID {
				sendJSON(w, http.StatusUnauthorized, false, "You're not authorized to modify this transfer", nil)
				return
			}
		} else {
			expiryDates := r.MultipartForm.Value["expiry-date"]
			if len(expiryDates) != 1 {
				sendJSON(w, http.StatusBadRequest, false, "Expected a transfer id or expiry date", nil)
				return
			}

			expiryDate, err := time.Parse("2006-02-03", expiryDates[0])
			if err != nil {
				slog.Error("Failed parsing date", "error", err, "input", expiryDates[0])
				sendJSON(w, http.StatusBadRequest, false, "Invalid date format, expected YYYY-MM-DD", nil)
				return
			}

			transfer = models.Transfer{
				UserID:     userID,
				ExpiryDate: expiryDate,
			}
			err = transfer.Create()
			if err != nil {
				slog.Error("Failed creating a new transfer", "error", err)
				sendJSON(w, http.StatusInternalServerError, false, "Failed creating new transfer", nil)
			}

			slog.Info("Created a new transfer")
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

		err = transfer.NewFile(fileHeader.Filename, int(fileHeader.Size))
		if err != nil {
			slog.Error("Failed adding new file to transfer object", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Handle file failed", nil)
			return
		}

		slog.Debug("Successfully created new file", "user", userID, "transfer", transfer.ID, "file", fileHeader.Filename)
		err = sendRedirect(w, http.StatusSeeOther, "../../upload/"+transfer.ID, transfer.ID) // Redirect to `/upload/<transfer_id>`

		if err != nil {
			sendError(w, http.StatusInternalServerError, "Failed sending redirect")
		}
	}
}
