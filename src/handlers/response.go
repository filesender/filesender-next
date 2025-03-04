package handlers

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

var templatesFS embed.FS

// Standard API response format
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
func sendJSON(w http.ResponseWriter, status int, success bool, message string, data interface{}) {
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

	json.NewEncoder(w).Encode(response)
}

// Send a template response with the given data
func sendTemplate(w http.ResponseWriter, tmpl string, data any) {
	tmplPath := "templates/" + tmpl + ".html"
	basePath := "templates/base.html"

	t, err := template.ParseFS(templatesFS, basePath, tmplPath)
	if err != nil {
		log.Println("Error parsing template:", err)
		sendError(w, 500, "Error rendering page")
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		log.Println("Error executing template:", err)
		sendError(w, 500, "Error rendering page")
	}
}

// Send an error response
func sendError(w http.ResponseWriter, status int, message string) {
	http.Error(w, message, status)
}
