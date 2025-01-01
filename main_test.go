package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dhcgn/iot-ephemeral-value-store/storage"
)

func TestMainFunction(t *testing.T) {
	// Set up command line arguments
	os.Args = []string{"cmd", "-persist-values-for=1h", "-store=./testdata", "-port=8081"}

	// Mock functions to avoid actual server start and storage creation
	createStorage = func(storePath string, persistDuration time.Duration) storage.StorageInstance {
		return storage.NewInMemoryStorage()
	}
	listenAndServe = func(srv *http.Server) {
		// Create a test server
		ts := httptest.NewServer(srv.Handler)
		defer ts.Close()

		// Perform a test request
		resp, err := http.Get(ts.URL + "/")
		if err != nil {
			t.Fatalf("Failed to perform test request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status OK, got %v", resp.Status)
		}
	}

	// Call the main function
	main()
}
