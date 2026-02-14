package httphandler

import (
	"net/http"

	"github.com/dhcgn/iot-ephemeral-value-store/domain"
)

func (c Config) KeyPairHandler(w http.ResponseWriter, r *http.Request) {
	uploadKey := domain.GenerateRandomKey()
	downloadKey, err := domain.DeriveDownloadKey(uploadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Error deriving download key", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"upload-key":   domain.AddUploadPrefix(uploadKey),
		"download-key": domain.AddDownloadPrefix(downloadKey),
	}
	jsonResponse(w, response)
}
