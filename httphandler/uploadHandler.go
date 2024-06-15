package httphandler

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dhcgn/iot-ephemeral-value-store/domain"
	"github.com/gorilla/mux"
)

func (c Config) UploadAndPatchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uploadKey := vars["uploadKey"]
	path := vars["param"]

	c.handleUpload(w, r, uploadKey, path, true)
}

func (c Config) UploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uploadKey := vars["uploadKey"]

	c.handleUpload(w, r, uploadKey, "", false)
}

func (c Config) handleUpload(w http.ResponseWriter, r *http.Request, uploadKey, path string, isPatch bool) {
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

	// Collect parameters
	paramMap := collectParams(r.URL.Query())

	// Add timestamp to params
	addTimestamp(paramMap)

	// Handle data storage
	if err := c.handleDataStorage(downloadKey, paramMap, path, isPatch); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Construct and return response
	constructAndReturnResponse(w, r, downloadKey, paramMap)
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

func (c Config) handleDataStorage(downloadKey string, paramMap map[string]string, path string, isPatch bool) error {
	var dataToStore map[string]interface{}

	if isPatch {
		existingData, err := c.StorageInstance.RetrieveExistingData(downloadKey)
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

	return c.StorageInstance.StoreData(downloadKey, dataToStore)
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
