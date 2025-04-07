package auth

import (
	"net/http"
)

// DummyAuth provides a hardcoded user for development/testing.
type DummyAuth struct{}

// UserAuth authenticates user
func (s *DummyAuth) UserAuth(_ *http.Request) (string, error) {
	return "dev", nil
}
