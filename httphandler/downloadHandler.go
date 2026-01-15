package httphandler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

func (c Config) DownloadPlainHandler(w http.ResponseWriter, r *http.Request) {
	c.downloadPlainHandler(w, r, false)
}

func (c Config) downloadPlainHandler(w http.ResponseWriter, r *http.Request, base64mode bool) {
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

	// If base64 mode is enabled, decode the value from base64url
	if base64mode {
		decoded, err := decodeBase64URL(value.(string))
		if err != nil {
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

func (c Config) DownloadBase64Handler(w http.ResponseWriter, r *http.Request) {
	c.downloadPlainHandler(w, r, true)
}

// DownloadRootHandler handles requests to /d/{downloadKey}/ and returns an HTML page
// with links to all available download endpoints
func (c Config) DownloadRootHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	downloadKey := vars["downloadKey"]

	jsonData, err := c.StorageInstance.GetJSON(downloadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Invalid download key or database error", http.StatusNotFound)
		return
	}

	// Parse the JSON data to extract field names
	paramMap := make(map[string]interface{})
	if err := json.Unmarshal(jsonData, &paramMap); err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
		return
	}

	c.StatsInstance.IncrementDownloads()

	// Generate HTML response with links
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Escape the downloadKey for safe HTML output
	escapedKey := html.EscapeString(downloadKey)
	
	// Start HTML
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Download Options</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        h1 { color: #333; }
        ul { list-style-type: none; padding: 0; }
        li { margin: 10px 0; }
        a { color: #0066cc; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .section { margin: 20px 0; }
    </style>
</head>
<body>
    <h1>Download Options</h1>
    <div class="section">
        <h2>JSON Format</h2>
        <ul>
            <li><a href="/d/%s/json">/d/%s/json</a></li>
        </ul>
    </div>
    <div class="section">
        <h2>Plain Text Fields</h2>
        <ul>
`, escapedKey, escapedKey)

	// Add links for each field in the JSON
	for key := range paramMap {
		// URL encode the key for use in href attribute
		urlEncodedKey := url.PathEscape(key)
		// HTML escape the key for display in text
		escapedFieldKey := html.EscapeString(key)
		fmt.Fprintf(w, `            <li><a href="/d/%s/plain/%s">/d/%s/plain/%s</a></li>
`, escapedKey, urlEncodedKey, escapedKey, escapedFieldKey)
	}

	// Close HTML
	fmt.Fprintf(w, `        </ul>
    </div>
</body>
</html>`)
}
