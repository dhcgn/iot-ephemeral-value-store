package httphandler

import (
	"fmt"
	"net/http"

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
	if err := domain.ValidateUploadKey(uploadKey); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Derive download key
	downloadKey, err := domain.DeriveDownloadKey(uploadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Error deriving download key", http.StatusInternalServerError)
		return
	}

	// Collect parameters
	paramMap := collectParams(r.URL.Query())

	// Add timestamp to params
	addTimestampToThisData(paramMap, path)

	// Handle data storage
	data, err := c.modifyData(downloadKey, paramMap, path, isPatch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.StorageInstance.Store(downloadKey, data)

	c.StatsInstance.IncrementUploads()

	// Construct and return response
	constructAndReturnResponse(w, r, downloadKey, paramMap)
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
