package httphandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func (c Config) PlainDownloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	downloadKey := vars["downloadKey"]
	param := vars["param"]

	jsonData, err := c.StorageInstance.GetJsonData(downloadKey)
	if err != nil {
		http.Error(w, "Invalid download key or database error", http.StatusNotFound)
		return
	}

	// Parse the JSON data to retrieve the specific parameter
	paramMap := make(map[string]interface{})
	if err := json.Unmarshal(jsonData, &paramMap); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
		return
	}

	// Split the param to get the keys for traversal
	keys := strings.Split(param, "/")

	// Traverse the map using the keys
	var value interface{} = paramMap
	for _, key := range keys {
		if m, ok := value.(map[string]interface{}); ok {
			value, ok = m[key]
			if !ok {
				http.Error(w, "Parameter not found", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "Invalid parameter path", http.StatusBadRequest)
			return
		}
	}

	// Return the value as plain text
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, value)
}

func (c Config) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	downloadKey := vars["downloadKey"]

	jsonData, err := c.StorageInstance.GetJsonData(downloadKey)
	if err != nil {
		http.Error(w, "Invalid download key or database error", http.StatusNotFound)
		return
	}

	// Set header and write the JSON data to the response writer
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
