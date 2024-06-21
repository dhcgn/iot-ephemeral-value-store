package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/httphandler"
	"github.com/dhcgn/iot-ephemeral-value-store/middleware"
	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
	"github.com/stretchr/testify/assert"
)

const (
	keyUp   = "8e88f1b62b946dd3fccfd8eaf54c9a2e5e27747c3662f2e20645073e4626d7c5"
	keyDown = "fcbbda7c04eba41d060b70d1bf7fde8c4a148a087729017d22fc54037c9eb11b"
)

type testCase struct {
	name               string
	url                string
	expectedStatusCode int
	checkBody          bool
	bodyContains       string
	bodyNotContains    string
}

func createTestEnvironment(t *testing.T) (*stats.Stats, httphandler.Config, middleware.Config) {
	stats := stats.NewStats()
	storageInMemory := storage.NewInMemoryStorage()

	httphandlerConfig := httphandler.Config{
		StorageInstance: storageInMemory,
		StatsInstance:   stats,
	}

	middlewareConfig := middleware.Config{
		RateLimitPerSecond: RateLimitPerSecond,
		RateLimitBurst:     RateLimitBurst,
		MaxRequestSize:     MaxRequestSize,
		StatsInstance:      stats,
	}

	return stats, httphandlerConfig, middlewareConfig
}

func runTests(t *testing.T, router http.Handler, tests []testCase) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.url, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			if tt.checkBody {
				if tt.bodyContains != "" {
					assert.Contains(t, rr.Body.String(), tt.bodyContains)
				}
				if tt.bodyNotContains != "" {
					assert.NotContains(t, rr.Body.String(), tt.bodyNotContains)
				}
			}
		})
	}
}

func buildURL(format string, a ...interface{}) string {
	return fmt.Sprintf(format, a...)
}

func TestCreateRouter(t *testing.T) {
	stats, httphandlerConfig, middlewareConfig := createTestEnvironment(t)
	router := createRouter(httphandlerConfig, middlewareConfig, stats)

	tests := []testCase{
		{"GET /", "/", http.StatusOK, false, "", ""},
		{"GET /kp", "/kp", http.StatusOK, false, "", ""},
		{"wrong upload key", "/wrong_upload_key", http.StatusBadRequest, false, "", ""},
	}

	runTests(t, router, tests)
}

func TestLegacyRoutes(t *testing.T) {
	stats, httphandlerConfig, middlewareConfig := createTestEnvironment(t)
	router := createRouter(httphandlerConfig, middlewareConfig, stats)

	tests := []testCase{
		{"Upload", buildURL("/%s/?value=8923423", keyUp), http.StatusOK, true, "Data uploaded successfully", ""},
		{"Download Plain", buildURL("/%s/plain/value", keyDown), http.StatusOK, true, "8923423\n", ""},
		{"Download JSON", buildURL("/%s/json", keyDown), http.StatusOK, true, "\"value\":\"8923423\"", ""},
	}

	runTests(t, router, tests)
}

func TestRoutesUploadDownload(t *testing.T) {
	stats, httphandlerConfig, middlewareConfig := createTestEnvironment(t)
	router := createRouter(httphandlerConfig, middlewareConfig, stats)

	tests := []testCase{
		{"Upload", buildURL("/u/%s/?value=8923423", keyUp), http.StatusOK, true, "Data uploaded successfully", ""},
		{"Download Plain", buildURL("/d/%s/plain/value", keyDown), http.StatusOK, true, "8923423\n", ""},
		{"Download JSON", buildURL("/d/%s/json", keyDown), http.StatusOK, true, "\"value\":\"8923423\"", ""},
	}

	runTests(t, router, tests)
}

func TestRoutesUploadDownloadDelete(t *testing.T) {
	stats, httphandlerConfig, middlewareConfig := createTestEnvironment(t)
	router := createRouter(httphandlerConfig, middlewareConfig, stats)

	tests := []testCase{
		{"Upload", buildURL("/u/%s/?value=8923423", keyUp), http.StatusOK, true, "Data uploaded successfully", ""},
		{"Download plain", buildURL("/d/%s/plain/value", keyDown), http.StatusOK, true, "8923423\n", ""},
		{"Download json", buildURL("/d/%s/json", keyDown), http.StatusOK, true, "\"value\":\"8923423\"", ""},
		{"Delete", buildURL("/delete/%s/", keyUp), http.StatusOK, true, "OK", ""},
		{"Download after delete plain", buildURL("/d/%s/plain/value", keyDown), http.StatusNotFound, false, "", ""},
		{"Download after delete json", buildURL("/d/%s/json", keyDown), http.StatusNotFound, false, "", ""},
	}

	runTests(t, router, tests)
}

func TestLegacyRoutesWithDifferentPathEndings(t *testing.T) {
	stats, httphandlerConfig, middlewareConfig := createTestEnvironment(t)
	router := createRouter(httphandlerConfig, middlewareConfig, stats)

	tests := []testCase{
		{"Upload Legacy with ending /", buildURL("/%s/?value=8923423", keyUp), http.StatusOK, true, "Data uploaded successfully", ""},
		{"Upload Legacy without ending /", buildURL("/%s?value=8923423", keyUp), http.StatusOK, true, "Data uploaded successfully", ""},
		{"Upload with ending /", buildURL("/u/%s/?value=8923423", keyUp), http.StatusOK, true, "Data uploaded successfully", ""},
		{"Upload without ending /", buildURL("/u/%s?value=8923423", keyUp), http.StatusOK, true, "Data uploaded successfully", ""},
	}

	runTests(t, router, tests)
}

func TestRoutesPatchDownload(t *testing.T) {
	stats, httphandlerConfig, middlewareConfig := createTestEnvironment(t)
	router := createRouter(httphandlerConfig, middlewareConfig, stats)

	tests := []testCase{
		{"Upload patch level 0", buildURL("/patch/%s/?value=1_4324232", keyUp), http.StatusOK, true, "Data uploaded successfully", ""},
		{"Upload patch level 0 no /", buildURL("/patch/%s?value_temp=1_4324232", keyUp), http.StatusOK, true, "Data uploaded successfully", ""},
		{"Upload patch level 2", buildURL("/patch/%s/1/2?value=2_8923423", keyUp), http.StatusOK, true, "Data uploaded successfully", ""},
		{"Download plain level 0", buildURL("/d/%s/plain/value", keyDown), http.StatusOK, true, "1_4324232\n", ""},
		{"Download plain level 2", buildURL("/d/%s/plain/1/2/value", keyDown), http.StatusOK, true, "2_8923423\n", ""},
		{"Download json level 0", buildURL("/d/%s/json", keyDown), http.StatusOK, true, "\"value\":\"1_4324232\"", ""},
		{"Download json level 2", buildURL("/d/%s/json", keyDown), http.StatusOK, true, "\"value\":\"2_8923423\"", ""},
		{"Download not contains empty key", buildURL("/d/%s/json", keyDown), http.StatusOK, false, "", "\"\""},
	}

	runTests(t, router, tests)
}
