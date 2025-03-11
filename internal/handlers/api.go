package handlers

import (
	"database/sql"
	"net/http"

	"codeberg.org/filesender/filesender-next/internal/models"
)

func CountFilesAPIHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		count, err := models.CountFiles(db)
		if err != nil {
			sendJSON(w, http.StatusInternalServerError, false, "Failed counting files in database", nil)
			return
		}

		sendJSON(w, http.StatusOK, true, "", map[string]int{"count": count})
	}
}
