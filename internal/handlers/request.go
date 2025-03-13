package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Receive JSON function, tries to unmarshal the request body
// If fails, returns a HTTP 400
func recvJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(v)
	if err != nil {
		slog.Error("Failed unmarshal JSON", "error", err)
		sendJSON(w, 400, false, "Incorrect JSON format", nil)
		return false
	}

	return true
}
