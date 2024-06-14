package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/dhcgn/iot-ephemeral-value-store/domain"
	"github.com/dhcgn/iot-ephemeral-value-store/httphandler"
	"github.com/dhcgn/iot-ephemeral-value-store/middleware"
	"github.com/gorilla/mux"
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

func initFlags() {
	myFlags := flag.NewFlagSet("iot-ephemeral-value-store", flag.ExitOnError)
	myFlags.StringVar(&persistDuration, "persist-values-for", DefaultPersistDuration, "Duration for which the values are stored before they are deleted.")
	myFlags.StringVar(&storePath, "store", DefaultStorePath, "Path to the directory where the values will be stored.")
	myFlags.IntVar(&port, "port", DefaultPort, "The port number on which the server will listen.")

	myFlags.Parse(os.Args[1:])
}

func main() {
	initFlags()

	db := createDatabase()
	defer db.Close()

	r := createRouter()

	serverAddress := fmt.Sprintf("127.0.0.1:%d", port)
	srv := &http.Server{
		Handler:      r,
		Addr:         serverAddress,
		WriteTimeout: WriteTimeout,
		ReadTimeout:  ReadTimeout,
	}

	fmt.Printf("Starting server on http://%s\n", serverAddress)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func createDatabase() *badger.DB {
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
	return db
}

func createRouter() *mux.Router {
	// Template parsing
	tmpl, err := template.ParseFS(staticFiles, "static/index.html")
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}

	r := mux.NewRouter()

	middleware := middleware.Config{
		RateLimitPerSecond: RateLimitPerSecond,
		RateLimitBurst:     RateLimitBurst,
		MaxRequestSize:     MaxRequestSize,
	}
	r.Use(middleware.EnableCORS)
	r.Use(middleware.LimitRequestSize)
	r.Use(middleware.RateLimit)

	httphandler := httphandler.Config{
		Db:              db,
		PersistDuration: persistDuration,
	}

	r.HandleFunc("/kp", httphandler.KeyPairHandler).Methods("GET")

	// Legacy routes
	r.HandleFunc("/{uploadKey}/", httphandler.UploadHandler).Methods("GET")
	r.HandleFunc("/{downloadKey}/json", httphandler.DownloadHandler).Methods("GET")
	r.HandleFunc("/{downloadKey}/plain/{param}", httphandler.PlainDownloadHandler).Methods("GET")

	// New routes
	r.HandleFunc("/u/{uploadKey}/", httphandler.UploadHandler).Methods("GET")
	r.HandleFunc("/d/{downloadKey}/json", httphandler.DownloadHandler).Methods("GET")
	r.HandleFunc("/d/{downloadKey}/plain/{param}", httphandler.PlainDownloadHandler).Methods("GET")
	// New routes with nestetd paths, eg. /u/1234/param1
	r.HandleFunc("/patch/{uploadKey}/{param:.*}", httphandler.UploadAndPatchHandler).Methods("GET")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := domain.GenerateRandomKey()
		data := PageData{
			UploadKey:      key,
			DownloadKey:    domain.DeriveDownloadKey(key),
			DataRentention: persistDuration,
		}
		tmpl.Execute(w, data)
	})

	staticSubFS, _ := fs.Sub(staticFiles, "static")
	r.PathPrefix("/").Handler(http.FileServer(http.FS(staticSubFS)))

	return r
}

type PageData struct {
	UploadKey      string
	DownloadKey    string
	DataRentention string
}
