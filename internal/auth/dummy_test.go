package auth_test

import (
	"net/http"
	"testing"

	"codeberg.org/filesender/filesender-next/internal/auth"
)

func TestDummyAuth(t *testing.T) {
	a := auth.DummyAuth{}
	userID, err := a.UserAuth(&http.Request{})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if userID != "dev" {
		t.Errorf("Expected user ID \"dev\", got: \"%s\"", userID)
	}
}
