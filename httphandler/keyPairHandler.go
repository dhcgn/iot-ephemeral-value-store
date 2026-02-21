package httphandler

import (
	"net/http"
)

func (c Config) KeyPairHandler(w http.ResponseWriter, r *http.Request) {
	uploadKey, downloadKey, err := c.DataService.GenerateKeyPair()
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Error generating key pair", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"upload-key":   uploadKey,
		"download-key": downloadKey,
	}
	jsonResponse(w, response)
}
