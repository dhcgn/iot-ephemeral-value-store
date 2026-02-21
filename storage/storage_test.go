// storage_test.go
package storage

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestInMemoryStorage(t *testing.T) {
	storage := NewInMemoryStorage()
	defer storage.Db.Close()

	// Test Store and Retrieve
	t.Run("Store and Retrieve", func(t *testing.T) {
		key := "test_key"
		data := map[string]interface{}{
			"name": "John Doe",
			"age":  30,
		}

		err := storage.Store(key, data)
		if err != nil {
			t.Fatalf("Failed to store data: %v", err)
		}

		retrieved, err := storage.Retrieve(key)
		if err != nil {
			t.Fatalf("Failed to retrieve data: %v", err)
		}

		n := retrieved["name"].(string)
		a := retrieved["age"].(float64)
		if n != "John Doe" || a != 30 {
			t.Errorf("Retrieved data does not match stored data. Got %v", retrieved)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		key := "delete_key"
		data := map[string]interface{}{"toDelete": true}

		err := storage.Store(key, data)
		if err != nil {
			t.Fatalf("Failed to store data: %v", err)
		}

		err = storage.Delete(key)
		if err != nil {
			t.Fatalf("Failed to delete data: %v", err)
		}

		// no error but empty data
		d, err := storage.Retrieve(key)
		if err != nil {
			t.Fatalf("Failed to retrieve data: %v", err)
		}
		if !reflect.DeepEqual(d, make(map[string]interface{})) {
			t.Errorf("Expected empty data, got %v", d)
		}
	})

	// Test TTL
	t.Run("TTL", func(t *testing.T) {
		key := "ttl_key"
		data := map[string]interface{}{"expiring": true}

		storage.PersistDuration = 1 * time.Second
		err := storage.Store(key, data)
		if err != nil {
			t.Fatalf("Failed to store data: %v", err)
		}

		time.Sleep(2 * time.Second)

		// no error but empty data
		d, err := storage.Retrieve(key)
		if err != nil {
			t.Fatalf("Failed to retrieve data: %v", err)
		}
		if !reflect.DeepEqual(d, make(map[string]interface{})) {
			t.Errorf("Expected empty data, got %v", d)
		}
	})
}

func TestStoreRawForTesting(t *testing.T) {
	storage := NewInMemoryStorage()
	defer storage.Db.Close()

	t.Run("StoreRawForTesting", func(t *testing.T) {
		key := "raw_test_key"
		rawData := []byte(`{"raw": "data"}`)

		err := storage.StoreRawForTesting(key, rawData)
		if err != nil {
			t.Fatalf("Failed to store raw data: %v", err)
		}

		retrieved, err := storage.Retrieve(key)
		if err != nil {
			t.Fatalf("Failed to retrieve data: %v", err)
		}

		if retrieved["raw"] != "data" {
			t.Errorf("Retrieved data does not match stored raw data. Got %v", retrieved)
		}
	})
}

func TestGetJSON(t *testing.T) {
	t.Run("GetJSON Success", func(t *testing.T) {
		storage := NewInMemoryStorage()
		defer storage.Db.Close()

		key := "json_test_key"
		data := map[string]interface{}{
			"field": "value",
		}

		err := storage.Store(key, data)
		if err != nil {
			t.Fatalf("Failed to store data: %v", err)
		}

		jsonData, err := storage.GetJSON(key)
		if err != nil {
			t.Fatalf("Failed to get JSON data: %v", err)
		}

		expectedJSON := `{"field":"value"}`
		if string(jsonData) != expectedJSON {
			t.Errorf("Expected JSON %s, got %s", expectedJSON, string(jsonData))
		}
	})

	t.Run("GetJSON Key Not Found", func(t *testing.T) {
		storage := NewInMemoryStorage()
		defer storage.Db.Close()

		key := "non_existent_key"

		_, err := storage.GetJSON(key)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}

		if err.Error() != "Key not found" {
			t.Errorf("Expected error 'Key not found', got %v", err)
		}
	})
}

func TestStoreJSONEncodingError(t *testing.T) {
	storage := NewInMemoryStorage()
	defer storage.Db.Close()

	t.Run("Store JSON Encoding Error", func(t *testing.T) {
		key := "json_encoding_error_key"
		data := map[string]interface{}{
			"invalid": make(chan int), // channels cannot be JSON encoded
		}

		err := storage.Store(key, data)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}

		if err.Error() != "error encoding data to JSON" {
			t.Errorf("Expected error 'error encoding data to JSON', got %v", err)
		}
	})
}

func TestNewPersistentStorage(t *testing.T) {
	dir, err := os.MkdirTemp("", "persistent-storage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	duration := 5 * time.Minute
	storage := NewPersistentStorage(dir, duration)
	defer storage.Db.Close()

	if storage.Db == nil {
		t.Fatal("Expected non-nil database")
	}
	if storage.PersistDuration != duration {
		t.Errorf("Expected persist duration %v, got %v", duration, storage.PersistDuration)
	}

	// Verify the storage is functional by storing and retrieving data
	key := "persistent_test_key"
	data := map[string]interface{}{"value": "42"}

	if err := storage.Store(key, data); err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	retrieved, err := storage.Retrieve(key)
	if err != nil {
		t.Fatalf("Failed to retrieve data: %v", err)
	}
	if retrieved["value"] != "42" {
		t.Errorf("Expected value=42, got %v", retrieved["value"])
	}
}

func TestNewPersistentStorage_createsDirectory(t *testing.T) {
	parent, err := os.MkdirTemp("", "persistent-storage-parent-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(parent) })

	// Use a subdirectory that does not yet exist
	storePath := parent + "/subdir/data"
	storage := NewPersistentStorage(storePath, time.Minute)
	defer storage.Db.Close()

	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		t.Errorf("Expected directory %q to be created", storePath)
	}
}
