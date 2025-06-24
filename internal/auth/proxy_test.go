package auth_test

import (
	"net/http"
	"strings"
	"testing"

	"codeberg.org/filesender/filesender-next/internal/auth"
)

func TestProxyAuth(t *testing.T) {
	a := auth.ProxyAuth{}

	t.Run("Invalid loopback", func(t *testing.T) {
		userID, err := a.UserAuth(&http.Request{
			RemoteAddr: "waow",
		})

		if err == nil {
			t.Errorf("Expected error, got nil")
		} else if !strings.Contains(err.Error(), "missing port in address") {
			t.Errorf("Expected error to be \"missing port in address\", got: \"%s\"", strings.Split(err.Error(), ": ")[1])
		}
		if userID != "" {
			t.Errorf("Expected user ID to be empty, got: \"%s\"", userID)
		}
	})

	t.Run("Not loopback", func(t *testing.T) {
		userID, err := a.UserAuth(&http.Request{
			RemoteAddr: "192.168.1.1:5678",
		})

		if err == nil {
			t.Errorf("Expected error, got nil")
		} else if !strings.Contains(err.Error(), "REMOTE_ADDR is NOT `localhost`") {
			t.Errorf("Expected error to be \"REMOTE_ADDR is NOT `localhost`\", got: \"%s\"", err.Error())
		}
		if userID != "" {
			t.Errorf("Expected user ID to be empty, got: \"%s\"", userID)
		}
	})

	t.Run("Remote user not set", func(t *testing.T) {
		userID, err := a.UserAuth(&http.Request{
			RemoteAddr: "127.0.0.1:5678",
		})

		if err == nil {
			t.Errorf("Expected error, got nil")
		} else if !strings.Contains(err.Error(), "HTTP header X-Remote-User is NOT set") {
			t.Errorf("Expected error to be \"HTTP header X-Remote-User is NOT set\", got: \"%s\"", err.Error())
		}
		if userID != "" {
			t.Errorf("Expected user ID to be empty, got: \"%s\"", userID)
		}
	})

	t.Run("Remote user empty", func(t *testing.T) {
		userID, err := a.UserAuth(&http.Request{
			RemoteAddr: "127.0.0.1:5678",
			Header: map[string][]string{
				"X-Remote-User": {""},
			},
		})

		if err == nil {
			t.Errorf("Expected error, got nil")
		} else if !strings.Contains(err.Error(), "HTTP header X-Remote-User is NOT set") {
			t.Errorf("Expected error to be \"HTTP header X-Remote-User is NOT set\", got: \"%s\"", err.Error())
		}
		if userID != "" {
			t.Errorf("Expected user ID to be empty, got: \"%s\"", userID)
		}
	})

	t.Run("Success", func(t *testing.T) {
		userID, err := a.UserAuth(&http.Request{
			RemoteAddr: "127.0.0.1:5678",
			Header: map[string][]string{
				"X-Remote-User": {"dev"},
			},
		})

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if userID != "dev" {
			t.Errorf("Expected user ID to be \"dev\", got: \"%s\"", userID)
		}
	})
}
