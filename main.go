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

// Set in build time
var (
	Version   string = "dev"
	BuildTime string = "unknown"
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

	httphandlerConfig := httphandler.Config{
		Db:              db,
		PersistDuration: persistDuration,
	}

	middlewareConfig := middleware.Config{
		RateLimitPerSecond: RateLimitPerSecond,
		RateLimitBurst:     RateLimitBurst,
		MaxRequestSize:     MaxRequestSize,
	}

	r := createRouter(httphandlerConfig, middlewareConfig)

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

func createRouter(hhc httphandler.Config, mc middleware.Config) *mux.Router {
	// Template parsing
	tmpl, err := template.ParseFS(staticFiles, "static/index.html")
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}

	r := mux.NewRouter()

	r.Use(mc.EnableCORS)
	r.Use(mc.LimitRequestSize)
	r.Use(mc.RateLimit)

	r.HandleFunc("/kp", hhc.KeyPairHandler).Methods("GET")

	// Legacy routes
	r.HandleFunc("/{uploadKey}/", hhc.UploadHandler).Methods("GET")
	r.HandleFunc("/{downloadKey}/json", hhc.DownloadHandler).Methods("GET")
	r.HandleFunc("/{downloadKey}/plain/{param}", hhc.PlainDownloadHandler).Methods("GET")

	// New routes
	r.HandleFunc("/u/{uploadKey}/", hhc.UploadHandler).Methods("GET")
	r.HandleFunc("/d/{downloadKey}/json", hhc.DownloadHandler).Methods("GET")
	r.HandleFunc("/d/{downloadKey}/plain/{param:.*}", hhc.PlainDownloadHandler).Methods("GET")
	// New routes with nestetd paths, eg. /u/1234/param1
	r.HandleFunc("/patch/{uploadKey}/{param:.*}", hhc.UploadAndPatchHandler).Methods("GET")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := domain.GenerateRandomKey()
		key_down, err := domain.DeriveDownloadKey(key)
		if err != nil {
			http.Error(w, "Error deriving download key", http.StatusInternalServerError)
			return
		}
		data := PageData{
			UploadKey:     key,
			DownloadKey:   key_down,
			DataRetention: persistDuration,
			Version:       Version,
			BuildTime:     BuildTime,
		}
		tmpl.Execute(w, data)
	})

	// Not Found handler
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
	})

	staticSubFS, _ := fs.Sub(staticFiles, "static")
	r.PathPrefix("/").Handler(http.FileServer(http.FS(staticSubFS)))

	return r
}

type PageData struct {
	UploadKey     string
	DownloadKey   string
	DataRetention string
	Version       string
	BuildTime     string
}
