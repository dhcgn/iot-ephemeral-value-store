package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

func GenerateRandomKey() string {
	randomBytes := make([]byte, 256/8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic("Failed to generate random bytes: " + err.Error())
	}
	return hex.EncodeToString(randomBytes)
}

func DeriveDownloadKey(uploadKey string) (string, error) {
	if len(uploadKey) != 64 {
		return "", fmt.Errorf("Invalid upload key length: %d", len(uploadKey))
	}

	hash := sha256.Sum256([]byte(uploadKey))
	return hex.EncodeToString(hash[:]), nil
}

func ValidateUploadKey(uploadKey string) error {
	uploadKey = strings.ToLower(uploadKey)
	decoded, err := hex.DecodeString(uploadKey)
	if err != nil {
		return errors.New("uploadKey must be a 256 bit hex string")
	}
	if len(decoded) != 32 {
		return errors.New("uploadKey must be a 256 bit hex string")
	}
	return nil
}
