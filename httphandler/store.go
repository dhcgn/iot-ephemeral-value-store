package httphandler

import (
	"encoding/json"
	"errors"

	"github.com/dgraph-io/badger/v3"
)

func (c Config) getJsonData(downloadKey string) ([]byte, error) {
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

func (c Config) storeData(downloadKey string, dataToStore map[string]interface{}) error {
	updatedJSONData, err := json.Marshal(dataToStore)
	if err != nil {
		return errors.New("error encoding data to JSON")
	}

	return c.Db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(downloadKey), updatedJSONData).WithTTL(c.PersistDuration)
		return txn.SetEntry(e)
	})
}

func (c Config) retrieveExistingData(downloadKey string) (map[string]interface{}, error) {
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
