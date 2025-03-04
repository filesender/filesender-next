package config

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Attempts to find and load a .env file from current directory or one level up
func LoadEnv() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Println("Error getting working directory")
		return
	}

	paths := []string{
		filepath.Join(cwd, ".env"),               // Current directory
		filepath.Join(filepath.Dir(cwd), ".env"), // One directory up
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			err := loadEnvFile(path)
			if err != nil {
				log.Printf("Error loading .env file: %v", err)
			}
			return
		}
	}

	log.Println("No .env file found")
}

// Load an .env file into environment variables
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		value = strings.Trim(value, `"'`)
		os.Setenv(key, value)
	}

	return scanner.Err()
}
