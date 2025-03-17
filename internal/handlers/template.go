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
			"MinDate":     minDate.Format(time.DateOnly),
			"DefaultDate": defaultDate.Format(time.DateOnly),
			"MaxDate":     maxDate.Format(time.DateOnly),
		}

		sendTemplate(w, "upload", data)
	}
}
