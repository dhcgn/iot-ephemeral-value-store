package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/gorilla/mux"
)

var (
	persistDuration string
	storePath       string
	port            int
	db              *badger.DB
)

func init() {
	flag.StringVar(&persistDuration, "persist-values-for", "24h", "Duration for which the values are stored before they are deleted.")
	flag.StringVar(&storePath, "store", "./data", "Path to the directory where the values will be stored.")
	flag.IntVar(&port, "port", 8080, "The port number on which the server will listen.")
}

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

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func keyPairHandler(w http.ResponseWriter, r *http.Request) {
	uploadKey := generateRandomKey()
	downloadKey := deriveDownloadKey(uploadKey)

	response := map[string]string{
		"upload-key":   uploadKey,
		"download-key": downloadKey,
	}
	jsonResponse(w, response)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uploadKey := vars["uploadKey"]

	duration, err := time.ParseDuration(persistDuration)
	if err != nil {
		http.Error(w, "Invalid duration format", http.StatusBadRequest)
		return
	}

	params := r.URL.Query()
	for key, values := range params {
		if len(values) > 0 {
			err := db.Update(func(txn *badger.Txn) error {
				e := badger.NewEntry([]byte(uploadKey+"_"+key), []byte(values[0])).WithTTL(duration)
				return txn.SetEntry(e)
			})
			if err != nil {
				http.Error(w, "Failed to save to database", http.StatusInternalServerError)
				return
			}
		}
	}
	fmt.Fprintln(w, "OK")
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	downloadKey := vars["downloadKey"]

	found := false
	data := make(map[string]string)
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := downloadKey + "_"
		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			key := string(item.Key())[len(prefix):]
			err := item.Value(func(val []byte) error {
				data[key] = string(val)
				found = true
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !found {
		http.Error(w, "Invalid download key", http.StatusForbidden)
		return
	}
	jsonResponse(w, data)
}

func main() {
	flag.Parse()

	absStorePath, err := filepath.Abs(storePath)
	if err != nil {
		log.Fatalf("Error resolving store path: %s", err)
	}

	if _, err := os.Stat(absStorePath); os.IsNotExist(err) {
		err = os.MkdirAll(absStorePath, 0755)
		if err != nil {
			log.Fatalf("Error creating store directory: %s", err)
		}
	}

	db, err = badger.Open(badger.DefaultOptions(absStorePath))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/kp", keyPairHandler).Methods("GET")
	r.HandleFunc("/{uploadKey}/", uploadHandler).Methods("GET")
	r.HandleFunc("/{downloadKey}/json", downloadHandler).Methods("GET")

	serverAddress := fmt.Sprintf("127.0.0.1:%d", port)
	srv := &http.Server{
		Handler:      r,
		Addr:         serverAddress,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("Starting server on %s\n", serverAddress)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
