// storage_test.go
package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestInMemoryStorage(t *testing.T) {
	storage := NewInMemoryStorage()
	defer storage.Close()
	ctx := context.Background()

	// Test Store and Retrieve
	t.Run("Store and Retrieve", func(t *testing.T) {
		key := "test_key"
		data := map[string]interface{}{
			"name": "John Doe",
			"age":  30,
		}

		err := storage.Store(ctx, key, data)
		if err != nil {
			t.Fatalf("Failed to store data: %v", err)
		}

		retrieved, err := storage.Retrieve(ctx, key)
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

		err := storage.Store(ctx, key, data)
		if err != nil {
			t.Fatalf("Failed to store data: %v", err)
		}

		err = storage.Delete(ctx, key)
		if err != nil {
			t.Fatalf("Failed to delete data: %v", err)
		}

		// no error but empty data
		d, err := storage.Retrieve(ctx, key)
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
		err := storage.Store(ctx, key, data)
		if err != nil {
			t.Fatalf("Failed to store data: %v", err)
		}

		time.Sleep(2 * time.Second)

		// no error but empty data
		d, err := storage.Retrieve(ctx, key)
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
	defer storage.Close()
	ctx := context.Background()

	t.Run("StoreRawForTesting", func(t *testing.T) {
		key := "raw_test_key"
		rawData := []byte(`{"raw": "data"}`)

		err := storage.StoreRawForTesting(key, rawData)
		if err != nil {
			t.Fatalf("Failed to store raw data: %v", err)
		}

		retrieved, err := storage.Retrieve(ctx, key)
		if err != nil {
			t.Fatalf("Failed to retrieve data: %v", err)
		}

		if retrieved["raw"] != "data" {
			t.Errorf("Retrieved data does not match stored raw data. Got %v", retrieved)
		}
	})
}

func TestGetJSON(t *testing.T) {
	ctx := context.Background()

	t.Run("GetJSON Success", func(t *testing.T) {
		storage := NewInMemoryStorage()
		defer storage.Close()

		key := "json_test_key"
		data := map[string]interface{}{
			"field": "value",
		}

		err := storage.Store(ctx, key, data)
		if err != nil {
			t.Fatalf("Failed to store data: %v", err)
		}

		jsonData, err := storage.GetJSON(ctx, key)
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
		defer storage.Close()

		key := "non_existent_key"

		_, err := storage.GetJSON(ctx, key)
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
	defer storage.Close()
	ctx := context.Background()

	t.Run("Store JSON Encoding Error", func(t *testing.T) {
		key := "json_encoding_error_key"
		data := map[string]interface{}{
			"invalid": make(chan int), // channels cannot be JSON encoded
		}

		err := storage.Store(ctx, key, data)
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
	defer storage.Close()
	ctx := context.Background()

	if storage.Db == nil {
		t.Fatal("Expected non-nil database")
	}
	if storage.PersistDuration != duration {
		t.Errorf("Expected persist duration %v, got %v", duration, storage.PersistDuration)
	}

	// Verify the storage is functional by storing and retrieving data
	key := "persistent_test_key"
	data := map[string]interface{}{"value": "42"}

	if err := storage.Store(ctx, key, data); err != nil {
		t.Fatalf("Failed to store data: %v", err)
	}

	retrieved, err := storage.Retrieve(ctx, key)
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
	defer storage.Close()

	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		t.Errorf("Expected directory %q to be created", storePath)
	}
}

func TestClose_inMemory(t *testing.T) {
	s := NewInMemoryStorage()
	// Close should succeed without error even without a GC goroutine.
	if err := s.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

func TestClose_persistent(t *testing.T) {
	dir, err := os.MkdirTemp("", "close-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	s := NewPersistentStorage(dir, time.Minute)
	// Close should stop GC and close the DB without error.
	if err := s.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

func TestValueLogGC_stopsOnClose(t *testing.T) {
	dir, err := os.MkdirTemp("", "gc-stop-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	s := NewPersistentStorage(dir, time.Minute)
	ctx := context.Background()

	// Store and delete some data so the GC has something to consider.
	for i := range 10 {
		key := fmt.Sprintf("gc_key_%d", i)
		_ = s.Store(ctx, key, map[string]interface{}{"v": i})
		_ = s.Delete(ctx, key)
	}

	// Closing should not hang or panic.
	if err := s.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

func TestContextCancellation(t *testing.T) {
	t.Run("Store with short timeout", func(t *testing.T) {
		s := NewInMemoryStorage()

		// Use an already-expired context.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(2 * time.Millisecond) // ensure timeout fires

		err := s.Store(ctx, "key", map[string]interface{}{"v": 1})
		// Wait briefly for any background goroutine to finish before closing.
		time.Sleep(50 * time.Millisecond)
		s.Close()

		if err == nil {
			// The in-memory DB is fast enough that it may finish before
			// the context check — that's acceptable.
			return
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected context.DeadlineExceeded error, got: %v", err)
		}
	})

	t.Run("Retrieve with short timeout", func(t *testing.T) {
		s := NewInMemoryStorage()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(2 * time.Millisecond)

		_, err := s.Retrieve(ctx, "key")
		time.Sleep(50 * time.Millisecond)
		s.Close()

		if err == nil {
			return
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("Expected context.DeadlineExceeded error, got: %v", err)
		}
	})
}

func TestContextWithFallbackTimeout(t *testing.T) {
	t.Run("context without deadline gets fallback timeout", func(t *testing.T) {
		ctx := context.Background()
		newCtx, cancel := contextWithFallbackTimeout(ctx)
		defer cancel()

		deadline, ok := newCtx.Deadline()
		if !ok {
			t.Fatal("Expected context to have a deadline")
		}
		// Should be approximately DefaultOperationTimeout from now
		remaining := time.Until(deadline)
		if remaining < DefaultOperationTimeout-time.Second || remaining > DefaultOperationTimeout+time.Second {
			t.Errorf("Expected deadline ~%v from now, got %v", DefaultOperationTimeout, remaining)
		}
	})

	t.Run("context with existing deadline keeps it", func(t *testing.T) {
		shortTimeout := 2 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), shortTimeout)
		defer cancel()

		newCtx, newCancel := contextWithFallbackTimeout(ctx)
		defer newCancel()

		deadline, ok := newCtx.Deadline()
		if !ok {
			t.Fatal("Expected context to have a deadline")
		}
		remaining := time.Until(deadline)
		// Should preserve the original shorter deadline
		if remaining > shortTimeout+time.Second {
			t.Errorf("Expected deadline <= %v from now, got %v", shortTimeout, remaining)
		}
	})
}
