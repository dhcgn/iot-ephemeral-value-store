package httphandler

import (
	"encoding/json"
	"html"
	"net/http"
	"time"

	"github.com/dgraph-io/badger/v3"
)

type Config struct {
	Db              *badger.DB
	PersistDuration time.Duration
}

func sanitizeInput(input string) string {
	// Escapes HTML special characters like <, >, & and quotes
	return html.EscapeString(input)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
