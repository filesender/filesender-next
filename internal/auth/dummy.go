package auth

import (
	"net/http"
)

// DummyAuth contains... nothing?
type DummyAuth struct{}

// UserAuth authenticates user
func (s *DummyAuth) UserAuth(_ *http.Request) (string, error) {
	return "dev", nil
}
