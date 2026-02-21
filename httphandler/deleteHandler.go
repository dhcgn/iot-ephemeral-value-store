package httphandler

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (c Config) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uploadKey := vars["uploadKey"]

	_, err := c.DataService.Delete(uploadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}
