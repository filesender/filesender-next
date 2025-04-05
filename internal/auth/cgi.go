package auth

import (
	"errors"
	"net/http"
)

// CgiAuth contains... nothing?
type CgiAuth struct{}

// UserAuth authenticates user
func (s *CgiAuth) UserAuth(r *http.Request) (string, error) {
	remoteUser := r.Header.Get("REMOTE_USER")
	if remoteUser == "" {
		return "", errors.New("HTTP header REMOTE_USER is NOT set")
	}

	return remoteUser, nil
}
