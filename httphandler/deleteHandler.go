package httphandler

import (
	"net/http"

	"github.com/dhcgn/iot-ephemeral-value-store/domain"
	"github.com/gorilla/mux"
)

func (c Config) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uploadKey := vars["uploadKey"]

	downloadKey, err := domain.DeriveDownloadKey(uploadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, "Error deriving download key", http.StatusInternalServerError)
		return
	}

	// TODO return NOT FOUND if key does not exist?
	if err := c.StorageInstance.Delete(downloadKey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}
