package httphandler

import (
	"net/http"

	"github.com/dhcgn/iot-ephemeral-value-store/domain"
)

func (c Config) KeyPairHandler(w http.ResponseWriter, r *http.Request) {
	uploadKey := domain.GenerateRandomKey()
	downloadKey := domain.DeriveDownloadKey(uploadKey)

	response := map[string]string{
		"upload-key":   uploadKey,
		"download-key": downloadKey,
	}
	jsonResponse(w, response)
}
