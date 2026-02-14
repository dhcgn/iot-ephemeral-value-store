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

func TestStripUploadPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"With prefix", "u_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"},
		{"Without prefix", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripUploadPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("StripUploadPrefix() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestStripDownloadPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"With prefix", "d_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"},
		{"Without prefix", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripDownloadPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("StripDownloadPrefix() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAddUploadPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal key", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "u_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"},
		{"Empty string", "", "u_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddUploadPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("AddUploadPrefix() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAddDownloadPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal key", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "d_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"},
		{"Empty string", "", "d_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddDownloadPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("AddDownloadPrefix() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateUploadKeyWithPrefix(t *testing.T) {
	tests := []struct {
		name      string
		uploadKey string
		wantErr   bool
	}{
		{"Valid key with u_ prefix", "u_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", false},
		{"Valid key without prefix", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", false},
		{"Valid key with u_ prefix uppercase", "u_1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF", false},
		{"Invalid length with prefix", "u_1234567890abcdef", true},
		{"Non-hex characters with prefix", "u_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdeg", true},
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

func TestDeriveDownloadKeyWithPrefix(t *testing.T) {
	tests := []struct {
		name      string
		uploadKey string
		wantErr   bool
	}{
		{"Valid key with u_ prefix", "u_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", false},
		{"Valid key without prefix", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", false},
		{"Invalid length with prefix", "u_1234567890abcdef", true},
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

func TestDeriveDownloadKeyConsistencyWithPrefix(t *testing.T) {
	// Test that prefixed and non-prefixed versions of the same key produce the same download key
	baseKey := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	prefixedKey := "u_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	
	downloadKey1, err1 := DeriveDownloadKey(baseKey)
	if err1 != nil {
		t.Fatalf("DeriveDownloadKey for base key failed: %v", err1)
	}
	
	downloadKey2, err2 := DeriveDownloadKey(prefixedKey)
	if err2 != nil {
		t.Fatalf("DeriveDownloadKey for prefixed key failed: %v", err2)
	}
	
	if downloadKey1 != downloadKey2 {
		t.Error("DeriveDownloadKey produces different results for prefixed and non-prefixed versions of the same key")
	}
}
