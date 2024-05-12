package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var (
	store = make(map[string]map[string]string) // map[uploadKey]map[key]value
)

func generateRandomKey() string {
	randomBytes := make([]byte, 256/8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic("Failed to generate random bytes: " + err.Error())
	}

	return hex.EncodeToString(randomBytes)
}

func deriveDownloadKey(uploadKey string) string {
	hash := sha256.Sum256([]byte(uploadKey))
	return hex.EncodeToString(hash[:])
}

func keyPairHandler(w http.ResponseWriter, r *http.Request) {
	uploadKey := generateRandomKey()
	downloadKey := deriveDownloadKey(uploadKey)
	store[uploadKey] = make(map[string]string)

	response := map[string]string{
		"upload-key":   uploadKey,
		"download-key": downloadKey,
	}
	jsonResponse(w, response)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uploadKey := vars["uploadKey"]

	if _, ok := store[uploadKey]; !ok {
		http.Error(w, "Invalid upload key", http.StatusForbidden)
		return
	}

	params := r.URL.Query()
	for key, values := range params {
		if len(values) > 0 {
			store[uploadKey][key] = values[0]
		}
	}

	fmt.Fprintln(w, "OK")
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	downloadKey := vars["downloadKey"]

	for uploadKey, data := range store {
		if deriveDownloadKey(uploadKey) == downloadKey {
			jsonResponse(w, data)
			return
		}
	}

	http.Error(w, "Invalid download key", http.StatusForbidden)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/kp", keyPairHandler).Methods("GET")
	r.HandleFunc("/{uploadKey}/", uploadHandler).Methods("GET")
	r.HandleFunc("/{downloadKey}/json", downloadHandler).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Starting server on :8080")
	srv.ListenAndServe()
}
