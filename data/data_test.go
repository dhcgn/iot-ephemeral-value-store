package data

import (
	"encoding/json"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/domain"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
)

func newTestService() (*Service, storage.StorageInstance) {
	s := storage.NewInMemoryStorage()
	return &Service{StorageInstance: s}, s
}

func TestGenerateKeyPair(t *testing.T) {
	svc, _ := newTestService()

	uploadKey, downloadKey, err := svc.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if uploadKey == "" {
		t.Error("Expected non-empty upload key")
	}
	if downloadKey == "" {
		t.Error("Expected non-empty download key")
	}
	// Download key should be derivable from upload key
	derived, _ := domain.DeriveDownloadKey(uploadKey)
	if derived != downloadKey {
		t.Errorf("Expected download key %s, got %s", derived, downloadKey)
	}
}

func TestUpload(t *testing.T) {
	svc, si := newTestService()

	uploadKey := domain.GenerateRandomKey()
	params := map[string]string{"temp": "23.5", "humidity": "45"}

	downloadKey, data, err := svc.Upload(uploadKey, params)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if downloadKey == "" {
		t.Error("Expected non-empty download key")
	}
	if data["temp"] != "23.5" {
		t.Errorf("Expected temp=23.5, got %v", data["temp"])
	}
	if data["timestamp"] == nil {
		t.Error("Expected timestamp to be set")
	}

	// Verify stored in storage
	stored, err := si.GetJSON(downloadKey)
	if err != nil {
		t.Fatalf("Expected data in storage, got error: %v", err)
	}
	var storedMap map[string]interface{}
	json.Unmarshal(stored, &storedMap)
	if storedMap["temp"] != "23.5" {
		t.Errorf("Stored temp mismatch: %v", storedMap["temp"])
	}
}

func TestUpload_InvalidKey(t *testing.T) {
	svc, _ := newTestService()

	_, _, err := svc.Upload("invalid", map[string]string{"temp": "1"})
	if err == nil {
		t.Error("Expected error for invalid upload key")
	}
}

func TestPatch(t *testing.T) {
	svc, si := newTestService()

	uploadKey := domain.GenerateRandomKey()

	// First patch at room1
	downloadKey, _, err := svc.Patch(uploadKey, "room1", map[string]string{"temp": "20"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Second patch at room2
	_, _, err = svc.Patch(uploadKey, "room2", map[string]string{"temp": "22"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify merged
	stored, _ := si.GetJSON(downloadKey)
	var storedMap map[string]interface{}
	json.Unmarshal(stored, &storedMap)

	room1, ok := storedMap["room1"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected room1 to be a map")
	}
	if room1["temp"] != "20" {
		t.Errorf("Expected room1 temp=20, got %v", room1["temp"])
	}

	room2, ok := storedMap["room2"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected room2 to be a map")
	}
	if room2["temp"] != "22" {
		t.Errorf("Expected room2 temp=22, got %v", room2["temp"])
	}
}

func TestPatch_InvalidKey(t *testing.T) {
	svc, _ := newTestService()

	_, _, err := svc.Patch("invalid", "", map[string]string{"temp": "1"})
	if err == nil {
		t.Error("Expected error for invalid upload key")
	}
}

func TestDownloadJSON(t *testing.T) {
	svc, si := newTestService()

	uploadKey := domain.GenerateRandomKey()
	downloadKey, _ := domain.DeriveDownloadKey(uploadKey)
	si.Store(downloadKey, map[string]interface{}{"temp": "23"})

	jsonData, err := svc.DownloadJSON(downloadKey)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var m map[string]interface{}
	json.Unmarshal(jsonData, &m)
	if m["temp"] != "23" {
		t.Errorf("Expected temp=23, got %v", m["temp"])
	}
}

func TestDownloadJSON_NotFound(t *testing.T) {
	svc, _ := newTestService()

	_, err := svc.DownloadJSON("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent key")
	}
}

func TestDownloadField(t *testing.T) {
	svc, si := newTestService()

	downloadKey := "testkey"
	si.Store(downloadKey, map[string]interface{}{
		"temp": "23",
		"room": map[string]interface{}{
			"humidity": "45",
		},
	})

	t.Run("Root field", func(t *testing.T) {
		val, err := svc.DownloadField(downloadKey, "temp")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if val != "23" {
			t.Errorf("Expected 23, got %v", val)
		}
	})

	t.Run("Nested field", func(t *testing.T) {
		val, err := svc.DownloadField(downloadKey, "room/humidity")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if val != "45" {
			t.Errorf("Expected 45, got %v", val)
		}
	})

	t.Run("Non-existent field", func(t *testing.T) {
		_, err := svc.DownloadField(downloadKey, "nonexistent")
		if err == nil {
			t.Error("Expected error for non-existent field")
		}
	})
}

func TestDelete(t *testing.T) {
	svc, si := newTestService()

	uploadKey := domain.GenerateRandomKey()
	downloadKey, _ := domain.DeriveDownloadKey(uploadKey)
	si.Store(downloadKey, map[string]interface{}{"temp": "23"})

	retDownloadKey, err := svc.Delete(uploadKey)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if retDownloadKey != downloadKey {
		t.Errorf("Expected download key %s, got %s", downloadKey, retDownloadKey)
	}

	// Verify deleted
	_, err = si.GetJSON(downloadKey)
	if err == nil {
		t.Error("Expected data to be deleted")
	}
}

func TestDelete_InvalidKey(t *testing.T) {
	svc, _ := newTestService()

	_, err := svc.Delete("invalid")
	if err == nil {
		t.Error("Expected error for invalid upload key")
	}
}

func TestTraverseField(t *testing.T) {
	data := map[string]interface{}{
		"temp": "23",
		"room": map[string]interface{}{
			"humidity": "45",
			"nested": map[string]interface{}{
				"deep": "value",
			},
		},
	}

	t.Run("Root level", func(t *testing.T) {
		val, err := TraverseField(data, "temp")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if val != "23" {
			t.Errorf("Expected 23, got %v", val)
		}
	})

	t.Run("Nested level", func(t *testing.T) {
		val, err := TraverseField(data, "room/humidity")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if val != "45" {
			t.Errorf("Expected 45, got %v", val)
		}
	})

	t.Run("Deep nested", func(t *testing.T) {
		val, err := TraverseField(data, "room/nested/deep")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if val != "value" {
			t.Errorf("Expected value, got %v", val)
		}
	})

	t.Run("Not found", func(t *testing.T) {
		_, err := TraverseField(data, "missing")
		if err == nil {
			t.Error("Expected error")
		}
	})

	t.Run("Invalid path", func(t *testing.T) {
		_, err := TraverseField(data, "temp/invalid")
		if err == nil {
			t.Error("Expected error for traversing non-map value")
		}
	})
}
