package handlers

import (
	"net/http"
	"time"
)

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
