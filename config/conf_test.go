package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetConfigPaths(t *testing.T) {
	expectedPath := ""
	switch runtime.GOOS {
	case "windows":
		expectedPath = filepath.Join(os.Getenv("ProgramData"), "FileSender", "config.conf")
	case "darwin":
		expectedPath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "filesender", "config.conf")
	default:
		expectedPath = filepath.Join(os.Getenv("HOME"), ".config", "filesender", "config.conf")
	}

	actualPath := GetConfigPaths()
	if actualPath != expectedPath {
		t.Errorf("Expected %s, got %s", expectedPath, actualPath)
	}
}

func TestWriteDefaultConfig(t *testing.T) {
	tempDir := t.TempDir()
	dest := filepath.Join(tempDir, "config.conf")

	err := writeDefaultConfig(dest)
	if err != nil {
		t.Fatalf("Failed to write default config: %v", err)
	}

	_, err = os.Stat(dest)
	if os.IsNotExist(err) {
		t.Errorf("Default config file was not created at %s", dest)
	}
}

func TestLoadConfigFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test_config.conf")
	if err != nil {
		t.Fatalf("Failed to create temp config file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	content := `[section]
key value
another_key another_value`

	if _, err := tempFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write test config data: %v", err)
	}
	tempFile.Close()

	config, err := loadConfigFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config file: %v", err)
	}

	expected := map[string]map[string]string{
		"section": {
			"key":         "value",
			"another_key": "another_value",
		},
	}

	if len(config) != len(expected) || len(config["section"]) != len(expected["section"]) {
		t.Errorf("Config output mismatch: got %+v, expected %+v", config, expected)
	}
}

func TestLoadConfigCreatesDefault(t *testing.T) {
	tempDir := t.TempDir()

	os.Setenv("HOME", tempDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	_, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if _, err := os.Stat(GetConfigPaths()); os.IsNotExist(err) {
		t.Errorf("Expected default config to be created, but it wasn't")
	}
}
