package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	// UploadKeyPrefix is the optional prefix for upload keys
	UploadKeyPrefix = "u_"
	// DownloadKeyPrefix is the optional prefix for download keys
	DownloadKeyPrefix = "d_"
)

// StripUploadPrefix removes the optional "u_" prefix from an upload key
func StripUploadPrefix(key string) string {
	return strings.TrimPrefix(key, UploadKeyPrefix)
}

// StripDownloadPrefix removes the optional "d_" prefix from a download key
func StripDownloadPrefix(key string) string {
	return strings.TrimPrefix(key, DownloadKeyPrefix)
}

// AddUploadPrefix adds the "u_" prefix to an upload key
func AddUploadPrefix(key string) string {
	return UploadKeyPrefix + key
}

// AddDownloadPrefix adds the "d_" prefix to a download key
func AddDownloadPrefix(key string) string {
	return DownloadKeyPrefix + key
}

func GenerateRandomKey() string {
	randomBytes := make([]byte, 256/8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic("Failed to generate random bytes: " + err.Error())
	}
	return hex.EncodeToString(randomBytes)
}

func DeriveDownloadKey(uploadKey string) (string, error) {
	// Strip the optional "u_" prefix from the upload key
	uploadKey = StripUploadPrefix(uploadKey)

	if len(uploadKey) != 64 {
		return "", fmt.Errorf("Invalid upload key length: %d", len(uploadKey))
	}

	hash := sha256.Sum256([]byte(uploadKey))
	return hex.EncodeToString(hash[:]), nil
}

func ValidateUploadKey(uploadKey string) error {
	// Strip the optional "u_" prefix
	uploadKey = StripUploadPrefix(uploadKey)

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
