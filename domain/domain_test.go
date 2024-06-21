package domain

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerateRandomKey(t *testing.T) {
	key1 := GenerateRandomKey()
	key2 := GenerateRandomKey()

	if len(key1) != 64 {
		t.Errorf("Generated key length is incorrect. Expected 64, got %d", len(key1))
	}

	if key1 == key2 {
		t.Error("Two generated keys are identical, which is highly improbable")
	}

	_, err := hex.DecodeString(key1)
	if err != nil {
		t.Errorf("Generated key is not a valid hex string: %v", err)
	}
}

func TestDeriveDownloadKey(t *testing.T) {
	tests := []struct {
		name      string
		uploadKey string
		wantErr   bool
	}{
		{"Valid key", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", false},
		{"Invalid length", "1234567890abcdef", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downloadKey, err := DeriveDownloadKey(tt.uploadKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeriveDownloadKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(downloadKey) != 64 {
				t.Errorf("DeriveDownloadKey() returned key with incorrect length. Expected 64, got %d", len(downloadKey))
			}
		})
	}
}

func TestValidateUploadKey(t *testing.T) {
	tests := []struct {
		name      string
		uploadKey string
		wantErr   bool
	}{
		{"Valid key", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", false},
		{"Valid key with uppercase", "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF", false},
		{"Invalid length", "1234567890abcdef", true},
		{"Non-hex characters", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdeg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUploadKey(tt.uploadKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUploadKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeriveDownloadKeyConsistency(t *testing.T) {
	uploadKey := GenerateRandomKey()
	downloadKey1, err1 := DeriveDownloadKey(uploadKey)
	if err1 != nil {
		t.Fatalf("First DeriveDownloadKey call failed: %v", err1)
	}

	downloadKey2, err2 := DeriveDownloadKey(uploadKey)
	if err2 != nil {
		t.Fatalf("Second DeriveDownloadKey call failed: %v", err2)
	}

	if downloadKey1 != downloadKey2 {
		t.Error("DeriveDownloadKey is not consistent for the same input")
	}
}

func TestGenerateRandomKeyUniqueness(t *testing.T) {
	keys := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		key := GenerateRandomKey()
		if keys[key] {
			t.Error("Duplicate key generated")
		}
		keys[key] = true
	}
}

func TestValidateUploadKeyCaseInsensitivity(t *testing.T) {
	baseKey := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	variations := []string{
		strings.ToUpper(baseKey),
		strings.ToLower(baseKey),
		strings.ToTitle(baseKey),
	}

	for _, key := range variations {
		err := ValidateUploadKey(key)
		if err != nil {
			t.Errorf("ValidateUploadKey failed for case variation: %s", key)
		}
	}
}
