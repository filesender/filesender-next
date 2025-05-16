package handlers

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
)

var templatesFS embed.FS

// Init to receive the embedded templates
func Init(fs embed.FS) {
	templatesFS = fs
}

// Send a template response with the given data
func sendTemplate(w http.ResponseWriter, tmpl string, data any) {
	t, err := template.ParseFS(templatesFS, path.Join("templates", tmpl+".html"))
	if err != nil {
		slog.Error("Error parsing template", "error", err)
		sendError(w, 500, "Error rendering page")
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		slog.Error("Error executing template", "error", err)
		sendError(w, 500, "Error rendering page")
	}
}

// Send an error response
func sendError(w http.ResponseWriter, status int, message string) {
	slog.Error("Sending error to user", "status", status, "message", message)
	http.Error(w, message, status)
}

// Sends a redirect
func sendRedirect(w http.ResponseWriter, status int, location string, body string) error {
	w.Header().Add("Location", location)
	w.WriteHeader(status)

	_, err := w.Write([]byte(body))
	if err != nil {
		slog.Error("Failed writing empty response", "error", err)
		return err
	}

	return nil
}

// Send incomplete upload response
// Based on https://datatracker.ietf.org/doc/draft-ietf-httpbis-resumable-upload/
func sendIncompleteResponse(w http.ResponseWriter, appRoot string, fileID string, maxUploadSize int64, bytesReceived int64) {
	w.Header().Add("Upload-Draft-Interop-Version", "7")
	w.Header().Add("Location", filepath.Join(appRoot+"api/upload", fileID))
	w.Header().Add("Upload-Limit", strconv.FormatInt(maxUploadSize, 10))
	w.Header().Add("Upload-Offset", strconv.FormatInt(bytesReceived, 10))
	w.WriteHeader(http.StatusAccepted)
}
