package main

import (
	"crypto/rand"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

const (
	// Server configuration
	MaxRequestSize     = 1024 * 10 // 10 KB for request size limit
	RateLimitPerSecond = 10        // Requests per second
	RateLimitBurst     = 5         // Burst capability

	// Database and server paths
	DefaultStorePath       = "./data"
	DefaultPersistDuration = "24h"

	// Server timeout settings
	WriteTimeout = 15 * time.Second
	ReadTimeout  = 15 * time.Second

	// HTTP server configuration
	DefaultPort = 8080
)

//go:embed static/*
var staticFiles embed.FS

var (
	persistDuration string
	storePath       string
	port            int
	db              *badger.DB
)

func init() {
	flag.StringVar(&persistDuration, "persist-values-for", DefaultPersistDuration, "Duration for which the values are stored before they are deleted.")
	flag.StringVar(&storePath, "store", DefaultStorePath, "Path to the directory where the values will be stored.")
	flag.IntVar(&port, "port", DefaultPort, "The port number on which the server will listen.")
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

// ValidateUploadKey checks if the uploadKey is a valid 256 bit hex string
func ValidateUploadKey(uploadKey string) error {
	uploadKey = strings.ToLower(uploadKey) // make it case insensitive
	decoded, err := hex.DecodeString(uploadKey)
	if err != nil {
		return errors.New("uploadKey must be a 256 bit hex string")
	}
	if len(decoded) != 32 { // 256 bits = 32 bytes
		return errors.New("uploadKey must be a 256 bit hex string")
	}
	return nil
}

func limitRequestSize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request size is too large
		if r.ContentLength > MaxRequestSize {
			http.Error(w, "Request size is too large", http.StatusRequestEntityTooLarge)
			return
		}

		next.ServeHTTP(w, r)
	})
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

	err := ValidateUploadKey(uploadKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Derive the download key from the upload key
	downloadKey := deriveDownloadKey(uploadKey)

	// Parse the duration for setting TTL on data
	duration, err := time.ParseDuration(persistDuration)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid duration format: %v", err), http.StatusBadRequest)
		return
	}

	// Collect all parameters into a map
	params := r.URL.Query()
	paramMap := make(map[string]string)
	for key, values := range params {
		if len(values) > 0 {
			paramMap[key] = values[0]
		}
	}

	// Get the current timestamp
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Add the timestamp to the map
	paramMap["timestamp"] = timestamp

	// Convert the parameter map to a JSON string
	jsonData, err := json.Marshal(paramMap)
	if err != nil {
		http.Error(w, "Error encoding parameters to JSON", http.StatusInternalServerError)
		return
	}

	// Store the JSON data in the database using the download key
	err = db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(downloadKey), jsonData).WithTTL(duration)
		return txn.SetEntry(e)
	})
	if err != nil {
		http.Error(w, "Failed to save to database", http.StatusInternalServerError)
		return
	}

	// Construct URLs for each parameter for the plainDownloadHandler
	urls := make(map[string]string)
	for key := range params {
		urls[key] = fmt.Sprintf("http://%s/%s/plain/%s", r.Host, downloadKey, key)
	}

	// Construct the download URL
	downloadURL := fmt.Sprintf("http://%s/%s/json", r.Host, downloadKey)

	// Return the download URL in the response
	jsonResponse(w, map[string]interface{}{
		"message":        "Data uploaded successfully",
		"download_url":   downloadURL,
		"parameter_urls": urls,
	})
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	downloadKey := vars["downloadKey"]

	var jsonData []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(downloadKey))
		if err != nil {
			return err
		}
		jsonData, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		http.Error(w, "Invalid download key or database error", http.StatusNotFound)
		return
	}

	// Set header and write the JSON data to the response writer
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func plainDownloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	downloadKey := vars["downloadKey"]
	param := vars["param"]

	var jsonData []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(downloadKey))
		if err != nil {
			return err
		}
		jsonData, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		http.Error(w, "Data not found", http.StatusNotFound)
		return
	}

	// Parse the JSON data to retrieve the specific parameter
	paramMap := make(map[string]string)
	if err := json.Unmarshal(jsonData, &paramMap); err != nil {
		http.Error(w, "Error decoding JSON", http.StatusInternalServerError)
		return
	}

	// Retrieve the value for the requested parameter
	value, ok := paramMap[param]
	if !ok {
		http.Error(w, "Parameter not found", http.StatusNotFound)
		return
	}

	// Return the value as plain text
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, value)
}

var clients = make(map[string]*rate.Limiter)
var mtx sync.Mutex

func getLimiter(ip string) *rate.Limiter {
	mtx.Lock()
	defer mtx.Unlock()

	limiter, exists := clients[ip]
	if !exists {
		limiter = rate.NewLimiter(RateLimitPerSecond, RateLimitBurst)
		clients[ip] = limiter
	}

	return limiter
}

func rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		limiter := getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
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

	r.Use(limitRequestSize)
	r.Use(rateLimit)

	r.HandleFunc("/kp", keyPairHandler).Methods("GET")
	r.HandleFunc("/{uploadKey}/", uploadHandler).Methods("GET")
	r.HandleFunc("/{downloadKey}/json", downloadHandler).Methods("GET")
	r.HandleFunc("/{downloadKey}/plain/{param}", plainDownloadHandler).Methods("GET")

	staticSubFS, _ := fs.Sub(staticFiles, "static")
	r.PathPrefix("/").Handler(http.FileServer(http.FS(staticSubFS)))

	serverAddress := fmt.Sprintf("127.0.0.1:%d", port)
	srv := &http.Server{
		Handler:      r,
		Addr:         serverAddress,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("Starting server on http://%s\n", serverAddress)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
