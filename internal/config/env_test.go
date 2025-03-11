package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	someKey := GetEnv("SOME_KEY", "result")
	if someKey != "result" {
		t.Errorf("Expected 'result', got '%s'", someKey)
	}

	os.Setenv("SOME_KEY", "wawa")
	someKey = GetEnv("SOME_KEY", "result")
	if someKey != "wawa" {
		t.Errorf("Expected 'wawa', got '%s'", someKey)
	}
}
