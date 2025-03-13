package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"codeberg.org/filesender/filesender-next/internal/middlewares"
	"codeberg.org/filesender/filesender-next/internal/models"
)

func CreateTransferAPIHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authenticated, userID := middlewares.CookieAuth(r)
		if !authenticated {
			sendJSON(w, 401, false, "You're not authenticated", nil)
			return
		}

		// Read requets body
		var requestBody createTransferAPIRequest
		success := recvJSON(w, r, &requestBody)
		if !success {
			return
		}

		// Handle nil pointer
		var expiryDate time.Time
		if requestBody.ExpiryDate != nil {
			expiryDate = *requestBody.ExpiryDate
		}

		transfer, err := models.CreateTransfer(db, models.Transfer{
			UserID:     userID,
			Subject:    requestBody.Subject,
			Message:    requestBody.Message,
			ExpiryDate: expiryDate,
		})

		if err != nil {
			sendJSON(w, 500, false, "Failed creating transfer", nil)
			return
		}

		sendJSON(w, 201, true, "", createTransferAPIResponse{
			Transfer: *transfer,
		})
	}
}
