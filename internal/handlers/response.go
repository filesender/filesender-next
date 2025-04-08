package handlers

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"codeberg.org/filesender/filesender-next/internal/models"
)

var templatesFS embed.FS

// Response struct standardizes the JSON response data
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// Init to receive the embedded templates
func Init(fs embed.FS) {
	templatesFS = fs
}

// Send a JSON response with the given data
func sendJSON(w http.ResponseWriter, status int, success bool, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: success,
	}

	if success {
		if data != nil {
			response.Data = data
		}
	} else {
		response.Message = message
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error("Error sending JSON", "error", err)
	}
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

// Sends file
func sendFile(stateDir string, w http.ResponseWriter, file *models.File) {
	filePath := filepath.Join(stateDir, file.UserID, file.FileName)
	f, err := os.Open(filePath)
	if err != nil {
		sendError(w, http.StatusNotFound, "File not found") // This should never happen, this is already checked before...
		return
	}
	defer func() {
		err := f.Close()
		if err != nil {
			slog.Error("Failed closing file", "error", err)
		}
	}()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+file.FileName+`"`)
	w.Header().Set("Content-Length", strconv.Itoa(file.ByteSize))

	_, err = io.Copy(w, f)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Error while downloading the file")
	}
}
