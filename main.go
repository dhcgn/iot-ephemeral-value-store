package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dhcgn/iot-ephemeral-value-store/domain"
	"github.com/dhcgn/iot-ephemeral-value-store/httphandler"
	"github.com/dhcgn/iot-ephemeral-value-store/mcphandler"
	"github.com/dhcgn/iot-ephemeral-value-store/middleware"
	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
	"github.com/gorilla/mux"
)

const (
	// Server configuration
	MaxRequestSize     = 1024 * 10 // 10 KB for request size limit
	RateLimitPerSecond = 100       // Requests per second
	RateLimitBurst     = 10        // Burst capability

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
	persistDurationString string
	storePath             string
	port                  int
)

// Set in build time
var (
	Version   string = "dev"
	BuildTime string = "unknown"
	Commit    string = "unknown"
)

func initFlags() {
	myFlags := flag.NewFlagSet("iot-ephemeral-value-store", flag.ExitOnError)
	myFlags.StringVar(&persistDurationString, "persist-values-for", DefaultPersistDuration, "Duration for which the values are stored before they are deleted.")
	myFlags.StringVar(&storePath, "store", DefaultStorePath, "Path to the directory where the values will be stored.")
	myFlags.IntVar(&port, "port", DefaultPort, "The port number on which the server will listen.")

	myFlags.Parse(os.Args[1:])
}

var (
	createStorage = func(storePath string, persistDuration time.Duration) storage.StorageInstance {
		return storage.NewPersistentStorage(storePath, persistDuration)
	}
	listenAndServe = func(srv *http.Server) {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}
)

func main() {
	fmt.Println("Starting iot-ephemeral-value-store-server", Version, "Build:", BuildTime, "Commit:", Commit)
	fmt.Println("https://github.com/dhcgn/iot-ephemeral-value-store")
	fmt.Println("")

	initFlags()

	persistDuration, err := time.ParseDuration(persistDurationString)
	if err != nil {
		log.Fatalf("Failed to parse duration: %v", err)
	}

	stats := stats.NewStats()

	storage := createStorage(storePath, persistDuration)
	defer storage.Db.Close()

	httphandlerConfig := httphandler.Config{
		StorageInstance: storage,
		StatsInstance:   stats,
	}

	middlewareConfig := middleware.Config{
		RateLimitPerSecond: RateLimitPerSecond,
		RateLimitBurst:     RateLimitBurst,
		MaxRequestSize:     MaxRequestSize,
		StatsInstance:      stats,
	}

	r := createRouter(httphandlerConfig, middlewareConfig, stats)

	serverAddress := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Handler:      r,
		Addr:         serverAddress,
		WriteTimeout: WriteTimeout,
		ReadTimeout:  ReadTimeout,
	}

	fmt.Printf("Starting server on http://localhost:%v\n", port)
	listenAndServe(srv)
}

func createRouter(hhc httphandler.Config, mc middleware.Config, stats *stats.Stats) *mux.Router {
	// Template parsing
	tmpl, err := template.ParseFS(staticFiles, "static/index.html")
	if err != nil {
		log.Fatal("Error parsing templates:", err)
	}

	downloadTmpl, err := template.ParseFS(staticFiles, "static/download.html")
	if err != nil {
		log.Fatal("Error parsing download template:", err)
	}

	// Set the download template in the config
	hhc.DownloadTemplate = downloadTmpl

	r := mux.NewRouter()

	r.Use(mc.EnableCORS)
	r.Use(mc.LimitRequestSize)
	r.Use(mc.RateLimit)

	// MCP endpoint - create MCP server
	mcpConfig := mcphandler.Config{
		StorageInstance: hhc.StorageInstance,
		StatsInstance:   hhc.StatsInstance,
		ServerHost:      fmt.Sprintf("http://localhost:%d", port),
	}
	mcpServer, err := mcphandler.NewMCPServer(mcpConfig)
	if err != nil {
		log.Fatal("Error creating MCP server:", err)
	}
	r.HandleFunc("/mcp", mcpServer.ServeHTTP).Methods("GET", "POST")

	r.HandleFunc("/kp", hhc.KeyPairHandler).Methods("GET")

	// Static files that need explicit handling before legacy routes
	r.HandleFunc("/llm.txt", func(w http.ResponseWriter, r *http.Request) {
		content, err := staticFiles.ReadFile("static/llm.txt")
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(content)
	}).Methods("GET")

	// Viewer page
	r.HandleFunc("/viewer", viewerHandler())

	// Legacy routes
	r.HandleFunc("/{uploadKey}", hhc.UploadHandler).Methods("GET")
	r.HandleFunc("/{uploadKey}/", hhc.UploadHandler).Methods("GET")
	r.HandleFunc("/{downloadKey}/json", hhc.DownloadJsonHandler).Methods("GET")
	r.HandleFunc("/{downloadKey}/plain/{param}", hhc.DownloadPlainHandler).Methods("GET")

	// New routes
	r.HandleFunc("/u/{uploadKey}", hhc.UploadHandler).Methods("GET")
	r.HandleFunc("/u/{uploadKey}/", hhc.UploadHandler).Methods("GET")

	r.HandleFunc("/d/{downloadKey}/json", hhc.DownloadJsonHandler).Methods("GET")
	r.HandleFunc("/d/{downloadKey}/plain/{param:.*}", hhc.DownloadPlainHandler).Methods("GET")
	r.HandleFunc("/d/{downloadKey}/plain-from-base64url/{param:.*}", hhc.DownloadBase64Handler).Methods("GET")
	r.HandleFunc("/d/{downloadKey}/", hhc.DownloadRootHandler).Methods("GET")
	r.HandleFunc("/d/{downloadKey}", hhc.DownloadRootHandler).Methods("GET")
	// New routes with nested paths, eg. /u/1234/param1
	r.HandleFunc("/patch/{uploadKey}", hhc.UploadAndPatchHandler).Methods("GET")
	r.HandleFunc("/patch/{uploadKey}/{param:.*}", hhc.UploadAndPatchHandler).Methods("GET")

	// Admin
	r.HandleFunc("/delete/{uploadKey}", hhc.DeleteHandler).Methods("GET")
	r.HandleFunc("/delete/{uploadKey}/", hhc.DeleteHandler).Methods("GET")

	r.HandleFunc("/", templateHandler(tmpl, stats))

	// Not Found handler
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
	})

	staticSubFS, _ := fs.Sub(staticFiles, "static")
	r.PathPrefix("/").Handler(http.FileServer(http.FS(staticSubFS)))

	return r
}

func templateHandler(tmpl *template.Template, stats *stats.Stats) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := domain.GenerateRandomKey()
		key_down, err := domain.DeriveDownloadKey(key)
		if err != nil {
			http.Error(w, "Error deriving download key", http.StatusInternalServerError)
			return
		}
		data := PageData{
			UploadKey:     key,
			DownloadKey:   key_down,
			DataRetention: persistDurationString,
			Version:       Version,
			BuildTime:     BuildTime,

			Uptime: stats.GetUptime(),

			StateData: stats.GetCurrentStats(),
		}
		tmpl.Execute(w, data)
	}
}

func viewerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := staticFiles.ReadFile("static/viewer.html")
		if err != nil {
			http.Error(w, "Viewer page not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(content)
	}
}

type PageData struct {
	UploadKey     string
	DownloadKey   string
	DataRetention string
	Version       string
	BuildTime     string

	Uptime time.Duration

	StateData stats.StatsData
}
