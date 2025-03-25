// Package crypto contains hashing & enc/decrypting functions
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"

	"golang.org/x/crypto/salsa20"
)

const baseSecret = "this_is_a_32_byte_secret_key____"

func deriveKey(transferID string) *[32]byte {
	sum := sha256.Sum256([]byte(baseSecret + transferID))
	var key [32]byte
	copy(key[:], sum[:])
	return &key
}

func generateNonce() (*[8]byte, error) {
	var nonce [8]byte
	_, err := io.ReadFull(rand.Reader, nonce[:])
	if err != nil {
		return nil, err
	}
	return &nonce, nil
}

// EncryptUserAndTransferID encrypts the user ID using Salsa20
func EncryptUserAndTransferID(userID, transferID string) (string, error) {
	key := deriveKey(transferID)
	nonce, err := generateNonce()
	if err != nil {
		slog.Error("Failed to generate nonce", "error", err)
		return "", err
	}

	plain := []byte(userID)
	cipherText := make([]byte, len(plain))
	salsa20.XORKeyStream(cipherText, plain, nonce[:], key)

	final := make([]byte, 0, len(nonce)+len(cipherText))
	final = append(final, nonce[:]...)
	final = append(final, cipherText...)

	return base64.URLEncoding.EncodeToString(final), nil
}

// DecryptUserAndTransferID decrypts the string into User ID
func DecryptUserAndTransferID(encoded, transferID string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		slog.Error("Failed to decode base64", "error", err)
		return "", err
	}
	if len(data) < 8 {
		return "", fmt.Errorf("invalid ciphertext, too short")
	}

	var nonce [8]byte
	copy(nonce[:], data[:8])
	cipherText := data[8:]

	key := deriveKey(transferID)

	plain := make([]byte, len(cipherText))
	salsa20.XORKeyStream(plain, cipherText, nonce[:], key)

	return string(plain), nil
}
