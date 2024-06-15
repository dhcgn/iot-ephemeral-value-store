package storage

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v3"
)

type StorageInstance struct {
	Db              *badger.DB
	PersistDuration time.Duration
}

type Storage interface {
	GetJSON(downloadKey string) ([]byte, error)
	Delete(downloadKey string) error
	Store(downloadKey string, dataToStore map[string]interface{}) error
	Retrieve(downloadKey string) (map[string]interface{}, error)
}

func NewInMemoryStorage() StorageInstance {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		log.Fatal(err)
	}

	return StorageInstance{
		Db: db,
		// 1min
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

	db, err := badger.Open(badger.DefaultOptions(absStorePath))
	if err != nil {
		log.Fatal(err)
	}

	return StorageInstance{
		Db:              db,
		PersistDuration: persistDuration,
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
