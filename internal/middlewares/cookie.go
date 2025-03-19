package middlewares

import (
	"log/slog"
	"net/http"
)

type CookieAuth struct {
	CookieName string
}

func (s *CookieAuth) User(r *http.Request) (string, error) {
	val, err := r.Cookie(s.CookieName)
	if err != nil {
		slog.Error("Failed reading cookie", "error", err)
		return "", err
	}
	slog.Info("User Authenticated", "UserID", val.Value)

	return val.Value, nil
}
