package middlewares

import (
	"fmt"
	"log/slog"
	"net/http"
)

type HeaderAuth struct {
	HeaderName string
}

func (s *HeaderAuth) User(r *http.Request) (string, error) {
	remoteUser := r.Header.Get(s.HeaderName)
	if remoteUser == "" {
		return "", fmt.Errorf("header \"%s\" not set", s.HeaderName)
	}

	slog.Info("User Authenticated", "UserID", remoteUser)

	return remoteUser, nil
}
