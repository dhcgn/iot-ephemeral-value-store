package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/httphandler"
	"github.com/dhcgn/iot-ephemeral-value-store/middleware"
	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
	"github.com/stretchr/testify/assert"
)

func createTestEnvireonment(t *testing.T, stats *stats.Stats) (httphandler.Config, middleware.Config) {
	storageInMemory := storage.NewInMemoryStorage()

	var httphandlerConfig = httphandler.Config{
		StorageInstance: storageInMemory,
		StatsInstance:   stats,
	}

	var middlewareConfig = middleware.Config{
		RateLimitPerSecond: RateLimitPerSecond,
		RateLimitBurst:     RateLimitBurst,
		MaxRequestSize:     MaxRequestSize,
		StatsInstance:      stats,
	}

	return httphandlerConfig, middlewareConfig
}

var key_up = "8e88f1b62b946dd3fccfd8eaf54c9a2e5e27747c3662f2e20645073e4626d7c5"
var key_down = "fcbbda7c04eba41d060b70d1bf7fde8c4a148a087729017d22fc54037c9eb11b"

func TestCreateRouter(t *testing.T) {
	stats := stats.NewStats()
	httphandlerConfig, middlewareConfig := createTestEnvireonment(t, stats)
	router := createRouter(httphandlerConfig, middlewareConfig, stats)

	tests := []struct {
		name               string
		method             string
		url                string
		expectedStatusCode int
	}{
		{"GET /", "GET", "/", http.StatusOK},
		{"GET /kp", "GET", "/kp", http.StatusOK},
		{"wrong upload key", "GET", "/wrong_upload_key", http.StatusBadRequest},
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
	stats := stats.NewStats()
	httphandlerConfig, middlewareConfig := createTestEnvireonment(t, stats)
	router := createRouter(httphandlerConfig, middlewareConfig, stats)

	tests := []struct {
		name               string
		method             string
		url                string
		expectedStatusCode int
		checkBody          bool
		bodyContains       string
	}{
		{"Upload", "GET", "/" + key_up + "/" + "?value=8923423", http.StatusOK, true, "Data uploaded successfully"},
		{"Download Plain", "GET", "/" + key_down + "/" + "plain/value", http.StatusOK, true, "8923423\n"},
		{"Downlaod JSON", "GET", "/" + key_down + "/" + "json", http.StatusOK, true, "\"value\":\"8923423\""},
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
	stats := stats.NewStats()
	httphandlerConfig, middlewareConfig := createTestEnvireonment(t, stats)
	router := createRouter(httphandlerConfig, middlewareConfig, stats)

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

func TestRoutesUploadDownloadDelete(t *testing.T) {
	stats := stats.NewStats()
	httphandlerConfig, middlewareConfig := createTestEnvireonment(t, stats)
	router := createRouter(httphandlerConfig, middlewareConfig, stats)

	tests := []struct {
		name               string
		url                string
		expectedStatusCode int
		bodyContains       string
	}{
		{"Upload", "/u/" + key_up + "/" + "?value=8923423", http.StatusOK, "Data uploaded successfully"},
		{"Download plain", "/d/" + key_down + "/" + "plain/value", http.StatusOK, "8923423\n"},
		{"Download json", "/d/" + key_down + "/" + "json", http.StatusOK, "\"value\":\"8923423\""},
		{"Delete", "/delete/" + key_up + "/", http.StatusOK, "OK"},
		{"Download after delete plain", "/d/" + key_down + "/" + "plain/value", http.StatusNotFound, ""},
		{"Download after delete json", "/d/" + key_down + "/" + "json", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			if tt.bodyContains != "" {
				assert.Contains(t, rr.Body.String(), tt.bodyContains)
			}
		})
	}
}

func TestLegacyRoutesWithDifferentPathEndings(t *testing.T) {
	stats := stats.NewStats()
	httphandlerConfig, middlewareConfig := createTestEnvireonment(t, stats)
	router := createRouter(httphandlerConfig, middlewareConfig, stats)

	tests := []struct {
		name               string
		method             string
		url                string
		expectedStatusCode int
		bodyContains       string
	}{
		{"Upload Legacy with ending /", "GET", "/" + key_up + "/" + "?value=8923423", http.StatusOK, "Data uploaded successfully"},
		{"Upload Legacy without ending /", "GET", "/" + key_up + "?value=8923423", http.StatusOK, "Data uploaded successfully"},

		{"Upload with ending /", "GET", "/u/" + key_up + "/" + "?value=8923423", http.StatusOK, "Data uploaded successfully"},
		{"Upload without ending /", "GET", "/u/" + key_up + "?value=8923423", http.StatusOK, "Data uploaded successfully"},
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
			if tt.bodyContains != "" {
				assert.Contains(t, rr.Body.String(), tt.bodyContains)
			}
		})
	}
}

func TestRoutesPatchDownload(t *testing.T) {
	stats := stats.NewStats()
	httphandlerConfig, middlewareConfig := createTestEnvireonment(t, stats)
	router := createRouter(httphandlerConfig, middlewareConfig, stats)

	tests := []struct {
		name               string
		method             string
		url                string
		expectedStatusCode int
		checkBody          bool
		bodyContains       string
		bodyNotContains    string
	}{
		{"Upload patch level 0", "GET", "/patch/" + key_up + "/" + "?value=1_4324232", http.StatusOK, true, "Data uploaded successfully", ""},
		{"Upload patch level 0 no /", "GET", "/patch/" + key_up + "?value_temp=1_4324232", http.StatusOK, true, "Data uploaded successfully", ""},
		{"Upload patch level 2", "GET", "/patch/" + key_up + "/1/2" + "?value=2_8923423", http.StatusOK, true, "Data uploaded successfully", ""},

		{"Download plain level 0", "GET", "/d/" + key_down + "/" + "plain/value", http.StatusOK, true, "1_4324232\n", ""},
		{"Download plain level 2", "GET", "/d/" + key_down + "/" + "plain/1/2/value", http.StatusOK, true, "2_8923423\n", ""},

		{"Download json level 0", "GET", "/d/" + key_down + "/" + "json", http.StatusOK, true, "\"value\":\"1_4324232\"", ""},
		{"Download json level 2", "GET", "/d/" + key_down + "/" + "json", http.StatusOK, true, "\"value\":\"2_8923423\"", ""},
		{"Download not contains empty key", "GET", "/d/" + key_down + "/" + "json", http.StatusOK, false, "", "\"\""},
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

			if tt.bodyNotContains != "" {
				assert.NotContains(t, rr.Body.String(), tt.bodyNotContains)
			}
		})
	}
}
