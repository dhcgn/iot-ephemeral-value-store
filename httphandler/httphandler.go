package httphandler

import (
	"encoding/json"
	"html"
	"net/http"

	"github.com/dhcgn/iot-ephemeral-value-store/storage"
)

type Config struct {
	StorageInstance storage.Storage
}

func sanitizeInput(input string) string {
	// Escapes HTML special characters like <, >, & and quotes
	return html.EscapeString(input)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
