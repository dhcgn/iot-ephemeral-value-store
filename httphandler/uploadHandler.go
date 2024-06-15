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

	err := ValidateUploadKey(uploadKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Derive the download key from the upload key
	downloadKey, err := domain.DeriveDownloadKey(uploadKey)
	if err != nil {
		http.Error(w, "Error deriving download key", http.StatusInternalServerError)
		return
	}

	// Parse the duration for setting TTL on data
	duration, err := time.ParseDuration(c.PersistDuration)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid duration format: %v", err), http.StatusBadRequest)
		return
	}

	// Collect all parameters into a map
	params := r.URL.Query()
	paramMap := make(map[string]string)
	for key, values := range params {
		if len(values) > 0 {
			// Sanitize each parameter value
			sanitizedValue := sanitizeInput(values[0])
			paramMap[key] = sanitizedValue
		}
	}

	// Get the current timestamp
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Add the timestamp to the map
	paramMap["timestamp"] = timestamp

	// Retrieve existing JSON data from the database
	var existingData map[string]interface{}
	err = c.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(downloadKey))
		if err != nil {
			// If the key does not exist, initialize an empty map
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
	if err != nil {
		http.Error(w, "Failed to retrieve existing data from database", http.StatusInternalServerError)
		return
	}

	// Merge the new data into the existing JSON structure
	mergeData(existingData, paramMap, strings.Split(path, "/"))

	// Convert the updated data to a JSON string
	updatedJSONData, err := json.Marshal(existingData)
	if err != nil {
		http.Error(w, "Error encoding updated data to JSON", http.StatusInternalServerError)
		return
	}

	// Store the updated JSON data in the database using the download key
	err = c.Db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(downloadKey), updatedJSONData).WithTTL(duration)
		return txn.SetEntry(e)
	})
	if err != nil {
		http.Error(w, "Failed to save updated data to database", http.StatusInternalServerError)
		return
	}

	// Construct URLs for each parameter for the plainDownloadHandler
	urls := make(map[string]string)
	for key := range params {
		urls[key] = fmt.Sprintf("http://%s/%s/plain/%s", r.Host, downloadKey, key)
	}

	// Construct the download URL
	downloadURL := fmt.Sprintf("http://%s/%s/json", r.Host, downloadKey)

	// Return the download URL in the response
	jsonResponse(w, map[string]interface{}{
		"message":        "Data uploaded successfully",
		"download_url":   downloadURL,
		"parameter_urls": urls,
	})
}

func (c Config) UploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uploadKey := vars["uploadKey"]

	err := ValidateUploadKey(uploadKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Derive the download key from the upload key
	downloadKey, err := domain.DeriveDownloadKey(uploadKey)
	if err != nil {
		http.Error(w, "Error deriving download key", http.StatusInternalServerError)
		return
	}

	// Parse the duration for setting TTL on data
	duration, err := time.ParseDuration(c.PersistDuration)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid duration format: %v", err), http.StatusBadRequest)
		return
	}

	// Collect all parameters into a map
	params := r.URL.Query()
	paramMap := make(map[string]string)
	for key, values := range params {
		if len(values) > 0 {
			// Sanitize each parameter value
			sanitizedValue := sanitizeInput(values[0])
			paramMap[key] = sanitizedValue
		}
	}

	// Get the current timestamp
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Add the timestamp to the map
	paramMap["timestamp"] = timestamp

	// Convert the parameter map to a JSON string
	jsonData, err := json.Marshal(paramMap)
	if err != nil {
		http.Error(w, "Error encoding parameters to JSON", http.StatusInternalServerError)
		return
	}

	// Store the JSON data in the database using the download key
	err = c.Db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(downloadKey), jsonData).WithTTL(duration)
		return txn.SetEntry(e)
	})
	if err != nil {
		http.Error(w, "Failed to save to database", http.StatusInternalServerError)
		return
	}

	// Construct URLs for each parameter for the plainDownloadHandler
	urls := make(map[string]string)
	for key := range params {
		urls[key] = fmt.Sprintf("http://%s/%s/plain/%s", r.Host, downloadKey, key)
	}

	// Construct the download URL
	downloadURL := fmt.Sprintf("http://%s/%s/json", r.Host, downloadKey)

	// Return the download URL in the response
	jsonResponse(w, map[string]interface{}{
		"message":        "Data uploaded successfully",
		"download_url":   downloadURL,
		"parameter_urls": urls,
	})
}

// mergeData merges the new data into the existing JSON structure based on the provided path.
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

// ValidateUploadKey checks if the uploadKey is a valid 256 bit hex string
func ValidateUploadKey(uploadKey string) error {
	uploadKey = strings.ToLower(uploadKey) // make it case insensitive
	decoded, err := hex.DecodeString(uploadKey)
	if err != nil {
		return errors.New("uploadKey must be a 256 bit hex string")
	}
	if len(decoded) != 32 { // 256 bits = 32 bytes
		return errors.New("uploadKey must be a 256 bit hex string")
	}
	return nil
}
