package id

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

func New() string {
	randomData := make([]byte, 16)
	rand.Read(randomData)

	return base64.RawURLEncoding.EncodeToString(randomData)
}

func Validate(encodedID string) error {
	// in order to validate the ID, we decode it and verify the length
	id, err := base64.RawURLEncoding.DecodeString(encodedID)
	if err != nil {
		return errors.New("invalid format")
	}
	if len(id) != 16 {
		return errors.New("invalid length")
	}

	return nil
}
