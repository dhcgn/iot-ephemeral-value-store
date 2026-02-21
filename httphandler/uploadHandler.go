package httphandler

import (
	"fmt"
	"net/http"

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
	paramMap := collectParams(r.URL.Query())

	// HTTP-specific: add per-key timestamps
	addTimestampToThisData(paramMap, path)

	var downloadKey string
	var err error

	if isPatch {
		downloadKey, _, err = c.DataService.Patch(uploadKey, path, paramMap)
	} else {
		downloadKey, _, err = c.DataService.Upload(uploadKey, paramMap)
	}

	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.StatsInstance.IncrementUploads()

	constructAndReturnResponse(w, r, downloadKey, paramMap)
}

func constructAndReturnResponse(w http.ResponseWriter, r *http.Request, downloadKey string, params map[string]string) {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	urls := make(map[string]string)
	for key := range params {
		urls[key] = fmt.Sprintf("%s://%s/d/%s/plain/%s", scheme, r.Host, downloadKey, key)
	}

	downloadURL := fmt.Sprintf("%s://%s/d/%s/json", scheme, r.Host, downloadKey)

	jsonResponse(w, map[string]interface{}{
		"message":        "Data uploaded successfully",
		"download_url":   downloadURL,
		"parameter_urls": urls,
	})
}
