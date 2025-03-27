// Package crypto contains hashing & enc/decrypting functions
package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"log/slog"
)

// HashToBase64 function hashes string input and returns base64 string
func HashToBase64(s string) (string, error) {
	h := sha256.New()
	_, err := h.Write([]byte(s))
	if err != nil {
		slog.Error("Failed writing into hash", "error", err)
		return "", err
	}

	bs := h.Sum(nil)
	encoded := base64.RawURLEncoding.EncodeToString(bs)

	return encoded, nil
}
