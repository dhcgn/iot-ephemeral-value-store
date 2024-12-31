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
