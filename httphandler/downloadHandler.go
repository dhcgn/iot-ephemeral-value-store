package httphandler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/dhcgn/iot-ephemeral-value-store/data"
	"github.com/gorilla/mux"
)

func (c Config) DownloadPlainHandler(w http.ResponseWriter, r *http.Request) {
	c.downloadPlainHandler(w, r, false)
}

func (c Config) downloadPlainHandler(w http.ResponseWriter, r *http.Request, base64mode bool) {
	vars := mux.Vars(r)
	downloadKey := vars["downloadKey"]
	param := vars["param"]

	jsonData, err := c.DataService.DownloadJSON(downloadKey)
	if err != nil {
		slog.Debug("download plain: failed to retrieve data", "error", err, "method", r.Method, "path", r.URL.Path)
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Invalid download key or database error", http.StatusNotFound)
		return
	}

	paramMap := make(map[string]interface{})
	if err := json.Unmarshal(jsonData, &paramMap); err != nil {
		slog.Error("download plain: failed to decode JSON", "error", err, "method", r.Method, "path", r.URL.Path)
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
		return
	}

	value, err := data.TraverseField(paramMap, param)
	if err != nil {
		slog.Debug("download plain: parameter not found", "error", err, "param", param, "method", r.Method, "path", r.URL.Path)
		c.StatsInstance.IncrementHTTPErrors()
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Parameter not found", http.StatusNotFound)
		} else {
			http.Error(w, "Invalid parameter path", http.StatusBadRequest)
		}
		return
	}

	// If base64 mode is enabled, decode the value from base64url
	if base64mode {
		decoded, err := decodeBase64URL(value.(string))
		if err != nil {
			slog.Error("download plain: failed to decode base64url", "error", err, "method", r.Method, "path", r.URL.Path)
			c.StatsInstance.IncrementHTTPErrors()
			http.Error(w, "Error decoding base64url", http.StatusInternalServerError)
			return
		}
		value = decoded
	}

	c.StatsInstance.IncrementDownloads()

	// Return the value as plain text
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, value)
}

func decodeBase64URL(encoded string) (string, error) {
	// Try decoding with standard base64
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err == nil {
		return string(decodedBytes), nil
	}

	// Try decoding with base64url
	decodedBytes, err = base64.URLEncoding.DecodeString(encoded)
	if err == nil {
		return string(decodedBytes), nil
	}

	// Try decoding with base64url without padding
	decodedBytes, err = base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}

func (c Config) DownloadJsonHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	downloadKey := vars["downloadKey"]

	jsonData, err := c.DataService.DownloadJSON(downloadKey)
	if err != nil {
		slog.Debug("download JSON: failed to retrieve data", "error", err, "method", r.Method, "path", r.URL.Path)
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Invalid download key or database error", http.StatusNotFound)
		return
	}

	c.StatsInstance.IncrementDownloads()

	// Set header and write the JSON data to the response writer
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (c Config) DownloadBase64Handler(w http.ResponseWriter, r *http.Request) {
	c.downloadPlainHandler(w, r, true)
}

// DownloadRootHandler handles requests to /d/{downloadKey}/ and returns an HTML page
// with links to all available download endpoints
func (c Config) DownloadRootHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	downloadKey := vars["downloadKey"]

	jsonData, err := c.DataService.DownloadJSON(downloadKey)
	if err != nil {
		slog.Debug("download root: failed to retrieve data", "error", err, "method", r.Method, "path", r.URL.Path)
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Invalid download key or database error", http.StatusNotFound)
		return
	}

	// Parse the JSON data to extract field names
	paramMap := make(map[string]interface{})
	if err := json.Unmarshal(jsonData, &paramMap); err != nil {
		slog.Error("download root: failed to decode JSON", "error", err, "method", r.Method, "path", r.URL.Path)
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
		return
	}

	c.StatsInstance.IncrementDownloads()

	// Prepare data for template
	type FieldData struct {
		URLEncoded string
		Name       string
	}

	// Collect all paths (including nested ones)
	paths := collectAllPaths(paramMap, "")

	// Sort paths alphabetically for consistent ordering
	sort.Strings(paths)

	var fields []FieldData
	if len(paths) > 0 {
		fields = make([]FieldData, 0, len(paths))
	}
	for _, path := range paths {
		fields = append(fields, FieldData{
			URLEncoded: url.PathEscape(path),
			Name:       path, // Template will auto-escape for display
		})
	}

	data := struct {
		DownloadKey string
		Fields      []FieldData
	}{
		DownloadKey: downloadKey, // Template will auto-escape
		Fields:      fields,
	}

	// Render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := c.DownloadTemplate.Execute(w, data); err != nil {
		slog.Error("download root: failed to render template", "error", err, "method", r.Method, "path", r.URL.Path)
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

// collectAllPaths recursively collects all paths in a nested map structure.
// The data parameter should be a map[string]interface{} representing the JSON structure.
// The prefix parameter should be an empty string ("") for root-level calls; it will be
// built up recursively as the function traverses nested maps.
// Returns a slice of path strings in the format "key" or "parent/child" for nested values.
func collectAllPaths(data interface{}, prefix string) []string {
	var paths []string

	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			var currentPath string
			if prefix == "" {
				currentPath = key
			} else {
				currentPath = prefix + "/" + key
			}

			// Check if the value is a nested map or a leaf value
			if nestedMap, ok := value.(map[string]interface{}); ok {
				// Recursively collect paths from nested maps
				nestedPaths := collectAllPaths(nestedMap, currentPath)
				paths = append(paths, nestedPaths...)
			} else {
				// This is a leaf value, add the path
				paths = append(paths, currentPath)
			}
		}
	}

	return paths
}
