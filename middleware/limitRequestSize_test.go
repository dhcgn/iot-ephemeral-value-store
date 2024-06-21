// limitRequestSize_test.go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/stats"
)

func TestLimitRequestSize(t *testing.T) {
	tests := []struct {
		name           string
		contentLength  int64
		maxRequestSize int64
		expectedStatus int
	}{
		{
			name:           "Request size within limit",
			contentLength:  1000,
			maxRequestSize: 2000,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Request size exceeds limit",
			contentLength:  3000,
			maxRequestSize: 2000,
			expectedStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:           "Request size equal to limit",
			contentLength:  2000,
			maxRequestSize: 2000,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStats := stats.NewStats()
			config := Config{
				MaxRequestSize: tt.maxRequestSize,
				StatsInstance:  mockStats,
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("POST", "/", nil)
			req.ContentLength = tt.contentLength

			rr := httptest.NewRecorder()

			middleware := config.LimitRequestSize(handler)
			middleware.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusRequestEntityTooLarge {
				if mockStats.GetCurrentStats().HTTPErrorCount != 1 {
					t.Errorf("IncrementHTTPErrors was not called")
				}
			} else {
				if mockStats.GetCurrentStats().HTTPErrorCount != 0 {
					t.Errorf("IncrementHTTPErrors was called unexpectedly")
				}
			}
		})
	}
}
