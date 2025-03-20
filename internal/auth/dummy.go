//go:build dev

package auth

import (
	"log/slog"
	"net/http"
)

func Auth(r *http.Request) (string, error) {
	return "dev", nil
}

func init() {
	slog.Info("dev mode is enabled, using dummy auth")
}
