package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// ErrStorageTimeout indicates a storage operation exceeded its deadline.
// The underlying write may still complete; callers must not assume the
// operation failed.
var ErrStorageTimeout = errors.New("storage operation timed out (commit status unknown)")

// ErrStorageDegraded indicates the storage backend is in a degraded state
// after a previous timeout, and new write operations are temporarily rejected.
var ErrStorageDegraded = errors.New("storage is degraded after a timeout, write rejected")

const (
	defaultWriteTimeout = 10 * time.Second
)

// badgerSlogLogger bridges badger's Logger interface to slog.
type badgerSlogLogger struct{}

func (badgerSlogLogger) Errorf(format string, args ...interface{}) {
	slog.Error(fmt.Sprintf("badger: "+format, args...))
}
func (badgerSlogLogger) Warningf(format string, args ...interface{}) {
	slog.Warn(fmt.Sprintf("badger: "+format, args...))
}
func (badgerSlogLogger) Infof(format string, args ...interface{}) {
	slog.Info(fmt.Sprintf("badger: "+format, args...))
}
func (badgerSlogLogger) Debugf(format string, args ...interface{}) {
	slog.Debug(fmt.Sprintf("badger: "+format, args...))
}

// StorageInstance wraps a BadgerDB database with write-timeout protection
// and a degraded-state circuit breaker.
type StorageInstance struct {
	Db              *badger.DB
	PersistDuration time.Duration
	WriteTimeout    time.Duration
	storePath       string

	// degraded is set to 1 after a write timeout to reject further writes
	// and prevent goroutine accumulation from a hung BadgerDB.
	degraded atomic.Int32
}

// Storage defines the data access interface for the value store.
type Storage interface {
	GetJSON(downloadKey string) ([]byte, error)
	Delete(downloadKey string) error
	Store(downloadKey string, dataToStore map[string]interface{}) error
	Retrieve(downloadKey string) (map[string]interface{}, error)
}

// HealthChecker reports the health of the storage backend.
type HealthChecker interface {
	CheckHealth() HealthStatus
}

// HealthStatus contains the result of a storage health check.
type HealthStatus struct {
	Healthy      bool   `json:"healthy"`
	Degraded     bool   `json:"degraded"`
	Message      string `json:"message,omitempty"`
	DiskFreeBytes int64 `json:"disk_free_bytes,omitempty"`
}

func NewInMemoryStorage() StorageInstance {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true).WithLogger(badgerSlogLogger{}))
	if err != nil {
		log.Fatal(err)
	}

	return StorageInstance{
		Db:           db,
		PersistDuration: 1 * time.Minute,
		WriteTimeout: defaultWriteTimeout,
	}
}

func NewPersistentStorage(storePath string, persistDuration time.Duration) StorageInstance {
	absStorePath, err := filepath.Abs(storePath)
	if err != nil {
		log.Fatalf("Error resolving store path: %s", err)
	}

	if _, err := os.Stat(absStorePath); os.IsNotExist(err) {
		err = os.MkdirAll(absStorePath, 0755)
		if err != nil {
			log.Fatalf("Error creating store directory: %s", err)
		}
	}

	db, err := badger.Open(badger.DefaultOptions(absStorePath).WithLogger(badgerSlogLogger{}))
	if err != nil {
		log.Fatal(err)
	}

	return StorageInstance{
		Db:              db,
		PersistDuration: persistDuration,
		WriteTimeout:    defaultWriteTimeout,
		storePath:       absStorePath,
	}
}

// updateWithTimeout runs a badger Update with a write timeout and
// circuit-breaker protection. If the operation times out, the storage
// is marked as degraded and subsequent writes are rejected.
func (c *StorageInstance) updateWithTimeout(fn func(txn *badger.Txn) error) error {
	if c.degraded.Load() != 0 {
		return ErrStorageDegraded
	}

	type result struct {
		err error
	}
	ch := make(chan result, 1)
	go func() {
		ch <- result{err: c.Db.Update(fn)}
	}()

	select {
	case res := <-ch:
		return res.err
	case <-time.After(c.WriteTimeout):
		c.degraded.Store(1)
		slog.Error("storage write timed out, marking storage as degraded",
			"timeout", c.WriteTimeout.String())
		return ErrStorageTimeout
	}
}

// ResetDegraded clears the degraded state, allowing writes again.
// This should be called after verifying the storage is healthy.
func (c *StorageInstance) ResetDegraded() {
	c.degraded.Store(0)
}

// IsDegraded reports whether the storage is in a degraded state.
func (c *StorageInstance) IsDegraded() bool {
	return c.degraded.Load() != 0
}

func (c *StorageInstance) GetJSON(downloadKey string) ([]byte, error) {
	var jsonData []byte
	err := c.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(downloadKey))
		if err != nil {
			return err
		}
		jsonData, err = item.ValueCopy(nil)
		return err
	})
	return jsonData, err
}

func (c *StorageInstance) Store(downloadKey string, dataToStore map[string]interface{}) error {
	updatedJSONData, err := json.Marshal(dataToStore)
	if err != nil {
		return errors.New("error encoding data to JSON")
	}

	return c.updateWithTimeout(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(downloadKey), updatedJSONData).WithTTL(c.PersistDuration)
		return txn.SetEntry(e)
	})
}

func (c *StorageInstance) Retrieve(downloadKey string) (map[string]interface{}, error) {
	var existingData map[string]interface{}
	err := c.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(downloadKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				existingData = make(map[string]interface{})
				return nil
			}
			return err
		}
		jsonData, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return json.Unmarshal(jsonData, &existingData)
	})
	return existingData, err
}

func (c *StorageInstance) Delete(downloadKey string) error {
	return c.updateWithTimeout(func(txn *badger.Txn) error {
		return txn.Delete([]byte(downloadKey))
	})
}

func (c *StorageInstance) StoreRawForTesting(downloadKey string, data []byte) error {
	return c.updateWithTimeout(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(downloadKey), data).WithTTL(c.PersistDuration)
		return txn.SetEntry(e)
	})
}

// CheckHealth performs a health check on the storage backend.
// It verifies the database is responsive, checks degraded state,
// and reports available disk space for persistent stores.
func (c *StorageInstance) CheckHealth() HealthStatus {
	if c.degraded.Load() != 0 {
		return HealthStatus{
			Healthy:  false,
			Degraded: true,
			Message:  "storage is degraded after a write timeout",
		}
	}

	// Verify database is responsive with a lightweight read.
	err := c.Db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("__healthcheck__"))
		if err == badger.ErrKeyNotFound {
			return nil
		}
		return err
	})
	if err != nil {
		return HealthStatus{
			Healthy: false,
			Message: fmt.Sprintf("database read check failed: %v", err),
		}
	}

	status := HealthStatus{
		Healthy: true,
		Message: "ok",
	}

	// Check disk space for persistent stores.
	if c.storePath != "" {
		free, err := diskFreeBytes(c.storePath)
		if err != nil {
			slog.Warn("failed to check disk space", "error", err)
		} else {
			status.DiskFreeBytes = free
			const minFreeBytes = 100 * 1024 * 1024 // 100 MB
			if free < minFreeBytes {
				status.Healthy = false
				status.Message = fmt.Sprintf("low disk space: %d bytes free", free)
			}
		}
	}

	return status
}
