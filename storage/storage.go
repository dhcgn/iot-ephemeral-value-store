package storage

import (
	"context"
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

// DefaultOperationTimeout is the maximum time a single BadgerDB transaction
// (View or Update) is allowed to run before the caller's context deadline
// takes precedence. If the caller already carries a shorter deadline (e.g.
// from the HTTP server's WriteTimeout) that deadline wins automatically.
const DefaultOperationTimeout = 10 * time.Second

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

<<<<<<< HEAD
// valueLogGCDiscardRatio is the ratio passed to RunValueLogGC.
// A value of 0.5 means Badger will rewrite a value log file if it can
// discard at least 50% of its space.
const valueLogGCDiscardRatio = 0.5

// StorageInstance wraps a BadgerDB instance with periodic value log garbage
// collection. Call Close to release resources and stop the GC goroutine.
type StorageInstance struct {
	Db              *badger.DB
	PersistDuration time.Duration
	stopGC          chan struct{}
}

// Storage defines the operations for persisting and retrieving ephemeral
// IoT data. All methods accept a context.Context so that callers (e.g. HTTP
// handlers) can propagate deadlines and cancellation to the underlying
// database operations.
=======
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
>>>>>>> origin/main
type Storage interface {
	GetJSON(ctx context.Context, downloadKey string) ([]byte, error)
	Delete(ctx context.Context, downloadKey string) error
	Store(ctx context.Context, downloadKey string, dataToStore map[string]interface{}) error
	Retrieve(ctx context.Context, downloadKey string) (map[string]interface{}, error)
}

// Close stops the periodic value log GC goroutine (if running) and closes the
// underlying BadgerDB.
func (c StorageInstance) Close() error {
	if c.stopGC != nil {
		close(c.stopGC)
	}
	return c.Db.Close()
}

// startValueLogGC starts a background goroutine that periodically runs
// Badger's value log garbage collection. The goroutine stops when stopCh
// is closed.
func startValueLogGC(db *badger.DB, interval time.Duration, stopCh chan struct{}) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				for {
					if err := db.RunValueLogGC(valueLogGCDiscardRatio); err != nil {
						break
					}
				}
			}
		}
	}()
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
<<<<<<< HEAD
		Db:              db,
		PersistDuration: time.Duration(1 * time.Minute),
=======
		Db:           db,
		PersistDuration: 1 * time.Minute,
		WriteTimeout: defaultWriteTimeout,
>>>>>>> origin/main
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

	stopCh := make(chan struct{})
	startValueLogGC(db, 5*time.Minute, stopCh)

	return StorageInstance{
		Db:              db,
		PersistDuration: persistDuration,
<<<<<<< HEAD
		stopGC:          stopCh,
	}
}

// viewWithContext runs a read-only BadgerDB transaction, honouring the
// caller's context deadline/cancellation. If ctx has no deadline, a
// fallback timeout of DefaultOperationTimeout is applied.
//
// When the context expires before the transaction completes, the function
// returns immediately with the context error. The background goroutine
// running the actual transaction will finish on its own — BadgerDB
// transactions are not cancellable, so we must let them complete to avoid
// corrupted state.
func (c StorageInstance) viewWithContext(ctx context.Context, fn func(txn *badger.Txn) error) error {
	ctx, cancel := contextWithFallbackTimeout(ctx)
	defer cancel()

	// Fast path: if context is already done, don't start a transaction.
	if ctx.Err() != nil {
		return fmt.Errorf("database read operation cancelled: %w", ctx.Err())
	}

	done := make(chan error, 1)
	go func() {
		done <- c.Db.View(fn)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		// The goroutine will still finish in the background. This is
		// intentional: BadgerDB transactions cannot be interrupted, and
		// the goroutine only reads from the DB.
		return fmt.Errorf("database read operation cancelled: %w", ctx.Err())
	}
}

// updateWithContext runs a read-write BadgerDB transaction, honouring the
// caller's context deadline/cancellation. If ctx has no deadline, a
// fallback timeout of DefaultOperationTimeout is applied.
func (c StorageInstance) updateWithContext(ctx context.Context, fn func(txn *badger.Txn) error) error {
	ctx, cancel := contextWithFallbackTimeout(ctx)
	defer cancel()

	if ctx.Err() != nil {
		return fmt.Errorf("database write operation cancelled: %w", ctx.Err())
	}

	done := make(chan error, 1)
	go func() {
		done <- c.Db.Update(fn)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("database write operation cancelled: %w", ctx.Err())
	}
}

// contextWithFallbackTimeout returns ctx with a DefaultOperationTimeout if
// ctx does not already carry a deadline. The caller must call the returned
// cancel function.
func contextWithFallbackTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, DefaultOperationTimeout)
}

func (c StorageInstance) GetJSON(ctx context.Context, downloadKey string) ([]byte, error) {
=======
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
>>>>>>> origin/main
	var jsonData []byte
	err := c.viewWithContext(ctx, func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(downloadKey))
		if err != nil {
			return err
		}
		jsonData, err = item.ValueCopy(nil)
		return err
	})
	return jsonData, err
}

<<<<<<< HEAD
func (c StorageInstance) Store(ctx context.Context, downloadKey string, dataToStore map[string]interface{}) error {
=======
func (c *StorageInstance) Store(downloadKey string, dataToStore map[string]interface{}) error {
>>>>>>> origin/main
	updatedJSONData, err := json.Marshal(dataToStore)
	if err != nil {
		return errors.New("error encoding data to JSON")
	}

<<<<<<< HEAD
	return c.updateWithContext(ctx, func(txn *badger.Txn) error {
=======
	return c.updateWithTimeout(func(txn *badger.Txn) error {
>>>>>>> origin/main
		e := badger.NewEntry([]byte(downloadKey), updatedJSONData).WithTTL(c.PersistDuration)
		return txn.SetEntry(e)
	})
}

<<<<<<< HEAD
func (c StorageInstance) Retrieve(ctx context.Context, downloadKey string) (map[string]interface{}, error) {
=======
func (c *StorageInstance) Retrieve(downloadKey string) (map[string]interface{}, error) {
>>>>>>> origin/main
	var existingData map[string]interface{}
	err := c.viewWithContext(ctx, func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(downloadKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
<<<<<<< HEAD
				// Return an empty map for missing keys so callers can
				// treat a non-existent key as an empty data set (e.g.
				// the first patch to a new key path).
=======
>>>>>>> origin/main
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

<<<<<<< HEAD
func (c StorageInstance) Delete(ctx context.Context, downloadKey string) error {
	return c.updateWithContext(ctx, func(txn *badger.Txn) error {
=======
func (c *StorageInstance) Delete(downloadKey string) error {
	return c.updateWithTimeout(func(txn *badger.Txn) error {
>>>>>>> origin/main
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
