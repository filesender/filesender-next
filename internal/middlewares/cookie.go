package middlewares

import (
	"log/slog"
	"net/http"
)

// Dummy function "cookie auth" just reading "session" cookie value
func CookieAuth(r *http.Request) (bool, string) {
	val, err := r.Cookie("session")
	if err == http.ErrNoCookie {
		return false, ""
	} else if err != nil {
		slog.Error("Failed reading cookie", "error", err)
		return false, ""
	}

	// For now just return the value, this is a dummy function anyways
	return true, val.Value
}
