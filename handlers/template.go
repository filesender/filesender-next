package handlers

import (
	"database/sql"
	"net/http"

	"codeberg.org/filesender/filesender-next/models"
)

func CountFilesTemplateHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		count, err := models.CountFiles(db)
		if err != nil {
			sendError(w, 500, "Failed counting files in database")
			return
		}

		data := map[string]any{
			"Title": "File count",
			"Count": count,
		}

		sendTemplate(w, "file-count", data)
	}
}
