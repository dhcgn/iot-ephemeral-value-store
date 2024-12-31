package httphandler

import (
	"encoding/json"
	"html"
	"net/http"
)

func sanitizeInput(input string) string {
	// Escapes HTML special characters like <, >, & and quotes
	return html.EscapeString(input)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

const upload_mode_key = "upload-mode"
const upload_mode_value_plain uploadMode = "plain"
const upload_mode_value_json uploadMode = "json"

type uploadMode string

func collectParams(params map[string][]string) map[string]string {
	paramMap := make(map[string]string)
	for key, values := range params {
		// Ignore the upload-mode key
		if key == upload_mode_key {
			continue
		}

		if len(values) > 0 {
			sanitizedValue := sanitizeInput(values[0])
			paramMap[key] = sanitizedValue
		}
	}
	return paramMap
}

func getUploadModeFromQuery(params map[string][]string) uploadMode {
	if values, exists := params[upload_mode_key]; exists {
		if len(values) > 0 {
			uploadMode := values[0]
			if uploadMode == string(upload_mode_value_plain) {
				return upload_mode_value_plain
			} else if uploadMode == string(upload_mode_value_json) {
				return upload_mode_value_json
			}
		}
	}
	return upload_mode_value_plain
}
