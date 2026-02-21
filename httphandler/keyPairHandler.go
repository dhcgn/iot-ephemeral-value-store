package httphandler

import (
	"log/slog"
	"net/http"

	"github.com/dhcgn/iot-ephemeral-value-store/domain"
)

func (c Config) KeyPairHandler(w http.ResponseWriter, r *http.Request) {
	uploadKey, downloadKey, err := c.DataService.GenerateKeyPair()
	if err != nil {
		slog.Error("keypair: failed to generate key pair", "error", err, "method", r.Method, "path", r.URL.Path)
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Error generating key pair", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"upload-key":   domain.AddUploadPrefix(uploadKey),
		"download-key": domain.AddDownloadPrefix(downloadKey),
	}
	jsonResponse(w, response)
}
