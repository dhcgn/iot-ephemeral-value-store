package httphandler

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/dhcgn/iot-ephemeral-value-store/domain"
	"github.com/gorilla/mux"
)

func (c Config) UploadAndPatchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uploadKey := vars["uploadKey"]
	path := vars["param"]

	handleUpload(w, r, c, uploadKey, path, true)
}

func (c Config) UploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uploadKey := vars["uploadKey"]

	handleUpload(w, r, c, uploadKey, "", false)
}

func handleUpload(w http.ResponseWriter, r *http.Request, c Config, uploadKey, path string, isPatch bool) {
	// Validate upload key
	if err := ValidateUploadKey(uploadKey); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Derive download key
	downloadKey, err := domain.DeriveDownloadKey(uploadKey)
	if err != nil {
		http.Error(w, "Error deriving download key", http.StatusInternalServerError)
		return
	}

	// Parse duration
	duration, err := parseDuration(c.PersistDuration)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid duration format: %v", err), http.StatusBadRequest)
		return
	}

	// Collect parameters
	paramMap := collectParams(r.URL.Query())

	// Add timestamp to params
	addTimestamp(paramMap)

	// Handle data storage
	if err := handleDataStorage(c, downloadKey, paramMap, path, isPatch, duration); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Construct and return response
	constructAndReturnResponse(w, r, downloadKey, paramMap)
}

func parseDuration(durationStr string) (time.Duration, error) {
	return time.ParseDuration(durationStr)
}

func collectParams(params map[string][]string) map[string]string {
	paramMap := make(map[string]string)
	for key, values := range params {
		if len(values) > 0 {
			sanitizedValue := sanitizeInput(values[0])
			paramMap[key] = sanitizedValue
		}
	}
	return paramMap
}

func addTimestamp(paramMap map[string]string) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	paramMap["timestamp"] = timestamp
}

func handleDataStorage(c Config, downloadKey string, paramMap map[string]string, path string, isPatch bool, duration time.Duration) error {
	var dataToStore map[string]interface{}

	if isPatch {
		existingData, err := retrieveExistingData(c, downloadKey)
		if err != nil {
			return err
		}
		mergeData(existingData, paramMap, strings.Split(path, "/"))
		dataToStore = existingData
	} else {
		dataToStore = make(map[string]interface{})
		for k, v := range paramMap {
			dataToStore[k] = v
		}
	}

	return storeData(c, downloadKey, dataToStore, duration)
}

func retrieveExistingData(c Config, downloadKey string) (map[string]interface{}, error) {
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

func storeData(c Config, downloadKey string, dataToStore map[string]interface{}, duration time.Duration) error {
	updatedJSONData, err := json.Marshal(dataToStore)
	if err != nil {
		return errors.New("error encoding data to JSON")
	}

	return c.Db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(downloadKey), updatedJSONData).WithTTL(duration)
		return txn.SetEntry(e)
	})
}

func constructAndReturnResponse(w http.ResponseWriter, r *http.Request, downloadKey string, params map[string]string) {
	urls := make(map[string]string)
	for key := range params {
		urls[key] = fmt.Sprintf("http://%s/d/%s/plain/%s", r.Host, downloadKey, key)
	}

	downloadURL := fmt.Sprintf("http://%s/d/%s/json", r.Host, downloadKey)

	jsonResponse(w, map[string]interface{}{
		"message":        "Data uploaded successfully",
		"download_url":   downloadURL,
		"parameter_urls": urls,
	})
}

func mergeData(existingData map[string]interface{}, newData map[string]string, path []string) {
	if len(path) == 0 || (len(path) == 1 && path[0] == "") {
		for k, v := range newData {
			existingData[k] = v
		}
		return
	}

	currentKey := path[0]
	if _, exists := existingData[currentKey]; !exists {
		existingData[currentKey] = make(map[string]interface{})
	}

	if nestedMap, ok := existingData[currentKey].(map[string]interface{}); ok {
		mergeData(nestedMap, newData, path[1:])
	} else {
		existingData[currentKey] = newData
	}
}

func ValidateUploadKey(uploadKey string) error {
	uploadKey = strings.ToLower(uploadKey)
	decoded, err := hex.DecodeString(uploadKey)
	if err != nil {
		return errors.New("uploadKey must be a 256 bit hex string")
	}
	if len(decoded) != 32 {
		return errors.New("uploadKey must be a 256 bit hex string")
	}
	return nil
}
