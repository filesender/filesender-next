package id_test

import (
	"strings"
	"testing"

	"codeberg.org/filesender/filesender-next/internal/id"
)

func TestNew(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		s, err := id.New()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(s) != 22 {
			t.Errorf("Expected random ID to have length of 22, got %d", len(s))
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("Invalid format", func(t *testing.T) {
		err := id.Validate("<")

		if err == nil {
			t.Errorf("Expected error, got nil")
		} else if !strings.Contains(err.Error(), "invalid format") {
			t.Errorf("Expected \"invalid format\" error, got \"%s\"", strings.Split(err.Error(), ": ")[0])
		}
	})

	t.Run("Invalid length", func(t *testing.T) {
		err := id.Validate("abc")

		if err == nil {
			t.Errorf("Expected error, got nil")
		} else if !strings.Contains(err.Error(), "invalid length") {
			t.Errorf("Expected \"invalid length\" error, got \"%s\"", strings.Split(err.Error(), ": ")[0])
		}
	})

	t.Run("Success", func(t *testing.T) {
		err := id.Validate("__ipzyw723kIA10r3uYWOg")

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})
}
