// Package hash contains hashing functions
package hash

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

const keyFileName = "hmac.key"

var hmacKey []byte

// Init function initialises hashing; generates key if not exists, or else reads from state dir
func Init(stateDir string) error {
	path := filepath.Join(stateDir, keyFileName)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		key := make([]byte, 32)

		_, err := rand.Read(key)
		if err != nil {
			slog.Error("Failed getting random key", "error", err)
			return fmt.Errorf("rand: %w", err)
		}

		err = os.WriteFile(path, key, 0o600)
		if err != nil {
			slog.Error("Failed writing key to file", "error", err)
			return fmt.Errorf("write key: %w", err)
		}

		hmacKey = key
		return nil
	}

	key, err := os.ReadFile(path)
	if err != nil {
		slog.Error("Failed reading key from file", "error", err)
		return fmt.Errorf("read key: %w", err)
	}

	hmacKey = key
	return nil
}

// ToBase64 function hashes string input and returns base64 string
func ToBase64(s string) (string, error) {
	mac := hmac.New(sha256.New, hmacKey)

	_, err := io.WriteString(mac, s)
	if err != nil {
		slog.Error("Failed writing into hash", "error", err)
		return "", err
	}

	sum := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sum), nil
}
