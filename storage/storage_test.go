// storage_test.go
package storage

import (
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
