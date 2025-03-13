package middlewares

import (
	"log/slog"
	"net/http"
)

// Dummy function "cookie auth" just reading "session" cookie value
func CookieAuth(r *http.Request) (authenticated bool, user_id string) {
	val, err := r.Cookie("session")
	if err == http.ErrNoCookie {
		return
	} else if err != nil {
		slog.Error("Failed reading cookie", "error", err)
		return
	}

	// For now just copy the value into "user_id", this is a dummy function anyways
	authenticated = true
	user_id = val.Value
	return
}
