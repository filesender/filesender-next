package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"codeberg.org/filesender/filesender-next/internal/models"
)

func CountFilesTemplateHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		count, err := models.CountFiles(db)
		if err != nil {
			sendError(w, 500, "Failed counting files in database")
			return
		}

		data := map[string]any{
			"Count": count,
		}

		sendTemplate(w, "file-count", data)
	}
}

func UploadTemplateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		minDate := time.Now().UTC().Add(time.Hour * 24)
		defaultDate := time.Now().UTC().Add(time.Hour * 24 * 7)
		maxDate := time.Now().UTC().Add(time.Hour * 24 * 30)

		data := map[string]any{
			"MinDate":     minDate.Format("2006-01-02"),
			"DefaultDate": defaultDate.Format("2006-01-02"),
			"MaxDate":     maxDate.Format("2006-01-02"),
		}

		sendTemplate(w, "upload", data)
	}
}
