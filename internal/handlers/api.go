package handlers

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"codeberg.org/filesender/filesender-next/internal/middlewares"
	"codeberg.org/filesender/filesender-next/internal/models"
)

func CreateTransferAPIHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authenticated, userID := middlewares.CookieAuth(r)
		if !authenticated {
			sendJSON(w, http.StatusUnauthorized, false, "You're not authenticated", nil)
			return
		}

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

		transfer, err := models.CreateTransfer(db, models.Transfer{
			UserID:     userID,
			Subject:    requestBody.Subject,
			Message:    requestBody.Message,
			ExpiryDate: expiryDate,
		})

		if err != nil {
			slog.Error("Failed creating transfer", "error", err)
			sendJSON(w, http.StatusInternalServerError, false, "Failed creating transfer", nil)
			return
		}

		slog.Debug("Successfully created new transfer", "user", userID, "transfer", transfer.ID)
		sendJSON(w, http.StatusCreated, true, "", createTransferAPIResponse{
			Transfer: *transfer,
		})
	}
}
