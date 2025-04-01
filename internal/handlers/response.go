package handlers

import (
	"archive/zip"
	"embed"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"path"

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

// Sends zipped file
func sendZippedFiles(w http.ResponseWriter, transfer *models.Transfer, files []string) {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="files.zip"`)

	zipWriter := zip.NewWriter(w)
	defer func() {
		err := zipWriter.Close()
		if err != nil {
			slog.Error("Failed closing zip writer", "error", err)
		}
	}()

	fileCount := 0
	for _, fname := range files {
		file, err := getFile(transfer, fname)
		if err != nil {
			slog.Error("Failed getting file", "file", fname, "error", err)
			continue
		}

		err = addFileToZip(zipWriter, file.Path)
		if err != nil {
			slog.Error("Error adding file to zip", "file", file.Path, "error", err)
			continue
		}

		file.DownloadCount++
		err = file.Save(transfer.UserID, transfer.ID, fname)
		if err != nil {
			slog.Error("Failed saving meta file", "file", file.Path, "error", err)
			// it wouldn't make sense to `continue` here
		}

		fileCount++
	}

	if fileCount == 0 {
		sendError(w, http.StatusInternalServerError, "No files available or could not be zipped")
	}
}
