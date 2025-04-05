package auth

import (
	"net/http"
)

type DummyAuth struct{}

func (s *DummyAuth) UserAuth(_ *http.Request) (string, error) {
	return "dev", nil
}
