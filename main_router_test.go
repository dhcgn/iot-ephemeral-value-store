package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/dhcgn/iot-ephemeral-value-store/httphandler"
	"github.com/dhcgn/iot-ephemeral-value-store/middleware"
	"github.com/stretchr/testify/assert"
)

func createTestEnvireonment(t *testing.T) (httphandler.Config, middleware.Config) {
	db := setupTestDB(t)
	var httphandlerConfig = httphandler.Config{
		Db:              db,
		PersistDuration: DefaultPersistDuration,
	}

	var middlewareConfig = middleware.Config{
		RateLimitPerSecond: RateLimitPerSecond,
		RateLimitBurst:     RateLimitBurst,
		MaxRequestSize:     MaxRequestSize,
	}

	return httphandlerConfig, middlewareConfig
}

var key_up = "8e88f1b62b946dd3fccfd8eaf54c9a2e5e27747c3662f2e20645073e4626d7c5"
var key_down = "fcbbda7c04eba41d060b70d1bf7fde8c4a148a087729017d22fc54037c9eb11b"

func TestCreateRouter(t *testing.T) {
	httphandlerConfig, middlewareConfig := createTestEnvireonment(t)

	router := createRouter(httphandlerConfig, middlewareConfig)

	tests := []struct {
		name               string
		method             string
		url                string
		expectedStatusCode int
	}{
		{"GET /", "GET", "/", http.StatusOK},
		{"GET /kp", "GET", "/kp", http.StatusOK},
		{"GET /nonexistent", "GET", "/nonexistent", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
		})
	}
}

func TestLegacyRoutes(t *testing.T) {
	httphandlerConfig, middlewareConfig := createTestEnvireonment(t)

	router := createRouter(httphandlerConfig, middlewareConfig)

	tests := []struct {
		name               string
		method             string
		url                string
		expectedStatusCode int
		checkBody          bool
		bodyContains       string
	}{
		{"GET /", "GET", "/" + key_up + "/" + "?value=8923423", http.StatusOK, true, "Data uploaded successfully"},
		{"GET /", "GET", "/" + key_down + "/" + "plain/value", http.StatusOK, true, "8923423\n"},
		{"GET /", "GET", "/" + key_down + "/" + "json", http.StatusOK, true, "\"value\":\"8923423\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			if tt.checkBody {
				assert.Contains(t, rr.Body.String(), tt.bodyContains)
			}
		})
	}
}

func TestRoutesUploadDownload(t *testing.T) {
	httphandlerConfig, middlewareConfig := createTestEnvireonment(t)

	router := createRouter(httphandlerConfig, middlewareConfig)

	tests := []struct {
		name               string
		method             string
		url                string
		expectedStatusCode int
		checkBody          bool
		bodyContains       string
	}{
		{"GET /", "GET", "/u/" + key_up + "/" + "?value=8923423", http.StatusOK, true, "Data uploaded successfully"},
		{"GET /", "GET", "/d/" + key_down + "/" + "plain/value", http.StatusOK, true, "8923423\n"},
		{"GET /", "GET", "/d/" + key_down + "/" + "json", http.StatusOK, true, "\"value\":\"8923423\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			if tt.checkBody {
				assert.Contains(t, rr.Body.String(), tt.bodyContains)
			}
		})
	}
}

func TestRoutesPatchDownload(t *testing.T) {
	httphandlerConfig, middlewareConfig := createTestEnvireonment(t)

	router := createRouter(httphandlerConfig, middlewareConfig)

	tests := []struct {
		name               string
		method             string
		url                string
		expectedStatusCode int
		checkBody          bool
		bodyContains       string
	}{
		{"GET patch", "GET", "/patch/" + key_up + "/1/" + "?value=8923423", http.StatusOK, true, "Data uploaded successfully"},
		//{"GET d plain", "GET", "/d/" + key_down + "/" + "plain/1/value", http.StatusOK, true, "8923423\n"},
		{"GET d json", "GET", "/d/" + key_down + "/" + "json", http.StatusOK, true, "\"value\":\"8923423\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			if tt.checkBody {
				assert.Contains(t, rr.Body.String(), tt.bodyContains)
			}
		})
	}
}

func setupTestDB(t *testing.T) *badger.DB {
	opts := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatalf("Failed to open Badger database: %v", err)
	}
	return db
}
