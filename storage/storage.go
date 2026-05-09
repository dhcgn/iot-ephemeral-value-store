package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v3"
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

type Storage interface {
	GetJSON(downloadKey string) ([]byte, error)
	Delete(downloadKey string) error
	Store(downloadKey string, dataToStore map[string]interface{}) error
	Retrieve(downloadKey string) (map[string]interface{}, error)
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

func NewInMemoryStorage() StorageInstance {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true).WithLogger(badgerSlogLogger{}))
	if err != nil {
		log.Fatal(err)
	}

	return StorageInstance{
		Db:              db,
		PersistDuration: time.Duration(1 * time.Minute),
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
		stopGC:          stopCh,
	}
}

func (c StorageInstance) GetJSON(downloadKey string) ([]byte, error) {
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

func (c StorageInstance) Store(downloadKey string, dataToStore map[string]interface{}) error {
	updatedJSONData, err := json.Marshal(dataToStore)
	if err != nil {
		return errors.New("error encoding data to JSON")
	}

	return c.Db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(downloadKey), updatedJSONData).WithTTL(c.PersistDuration)
		return txn.SetEntry(e)
	})
}

func (c StorageInstance) Retrieve(downloadKey string) (map[string]interface{}, error) {
	var existingData map[string]interface{}
	err := c.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(downloadKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				// TODO Why?
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

func (c StorageInstance) Delete(downloadKey string) error {
	return c.Db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(downloadKey))
	})
}

func (c StorageInstance) StoreRawForTesting(downloadKey string, data []byte) error {
	return c.Db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(downloadKey), data).WithTTL(c.PersistDuration)
		return txn.SetEntry(e)
	})
}
