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

	jsonData, err := c.StorageInstance.GetJSON(downloadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Invalid download key or database error", http.StatusNotFound)
		return
	}

	// Parse the JSON data to retrieve the specific parameter
	paramMap := make(map[string]interface{})
	if err := json.Unmarshal(jsonData, &paramMap); err != nil {
		c.StatsInstance.IncrementHTTPErrors()
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
				c.StatsInstance.IncrementHTTPErrors()
				http.Error(w, "Parameter not found", http.StatusNotFound)
				return
			}
		} else {
			c.StatsInstance.IncrementHTTPErrors()
			http.Error(w, "Invalid parameter path", http.StatusBadRequest)
			return
		}
	}

	c.StatsInstance.IncrementDownloads()

	// Return the value as plain text
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, value)
}

func (c Config) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	downloadKey := vars["downloadKey"]

	jsonData, err := c.StorageInstance.GetJSON(downloadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Invalid download key or database error", http.StatusNotFound)
		return
	}

	c.StatsInstance.IncrementDownloads()

	// Set header and write the JSON data to the response writer
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
