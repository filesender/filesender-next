package auth

import (
	"errors"
	"net/http"
)

// CgiAuth implements Auth using the REMOTE_USER heade
type CgiAuth struct{}

// UserAuth authenticates user
func (s *CgiAuth) UserAuth(r *http.Request) (string, error) {
	remoteUser := r.Header.Get("REMOTE_USER")
	if remoteUser == "" {
		return "", errors.New("HTTP header REMOTE_USER is NOT set")
	}

	return remoteUser, nil
}
