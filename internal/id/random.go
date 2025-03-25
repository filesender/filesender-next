package id

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

func New() string {
	randomData := make([]byte, 16)
	rand.Read(randomData)

	return hex.EncodeToString(randomData)
}

func Validate(in string) error {
	if len(in) != 32 {
		return errors.New("invalid length")
	}
	// there may be a better way to validate "hex" chars, but I am not sure
	// using a regexp will be better in this case. The hex.DecodeString in Go
	// is quite robust from what I can see in the code...
	_, err := hex.DecodeString(in)
	if err != nil {
		return errors.New("invalid format")
	}

	return nil
}
