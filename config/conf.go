package config

import (
	"bufio"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed default.conf
var defaultConf embed.FS

// Returns appropriate path to place config file
func GetConfigPaths() (path string) {
	switch runtime.GOOS {
	case "windows":
		path = filepath.Join(os.Getenv("ProgramData"), "FileSender", "config.conf")
	case "darwin": // macOS
		path = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "filesender", "config.conf")
	default: // Linux and other Unix-like OS
		path = filepath.Join(os.Getenv("HOME"), ".config", "filesender", "config.conf")
	}
	return
}

// Checks if config exists, if not, copies default config to default location
func LoadConfig() (map[string]map[string]string, error) {
	confPath := GetConfigPaths()

	if _, err := os.Stat(confPath); err != nil {
		err = writeDefaultConfig(confPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		log.Printf("Created default config at %s", confPath)
	}

	return loadConfigFile(confPath)
}

// Copies embedded default config to destination
func writeDefaultConfig(destination string) error {
	content, err := defaultConf.ReadFile("default.conf")
	if err != nil {
		return fmt.Errorf("failed to read embedded config: %w", err)
	}

	dir := filepath.Dir(destination)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return os.WriteFile(destination, content, 0644)
}

// Reads .conf file and returns key-value
func loadConfigFile(filename string) (map[string]map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := make(map[string]map[string]string)
	var currentSection string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = strings.TrimSpace(line[1 : len(line)-1])
			config[currentSection] = make(map[string]string)
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		key := parts[0]
		value := strings.Join(parts[1:], " ")

		if currentSection == "" {
			currentSection = "default"
			if _, exists := config[currentSection]; !exists {
				config[currentSection] = make(map[string]string)
			}
		}

		config[currentSection][key] = value
	}

	return config, scanner.Err()
}

// Deletes config file if exists
func DeleteConfigFile() error {
	confPath := GetConfigPaths()

	if _, err := os.Stat(confPath); err != nil {
		log.Printf("Config file does not exist")
		return nil
	}

	return os.Remove(confPath)
}
