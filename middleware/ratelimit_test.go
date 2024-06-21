package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/stats"
)

func TestRateLimit(t *testing.T) {
	tests := []struct {
		name            string
		remoteAddr      string
		requests        int
		expectRateLimit bool
	}{
		{
			name:            "Non-local IP should be rate limited",
			remoteAddr:      "192.168.1.1:1234",
			requests:        10,
			expectRateLimit: true,
		},
		{
			name:            "Local IP should not be rate limited",
			remoteAddr:      "127.0.0.1:1234",
			requests:        10,
			expectRateLimit: false,
		},
		{
			name:            "Empty remote addr should not be rate limited",
			remoteAddr:      "",
			requests:        10,
			expectRateLimit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStats := stats.NewStats()
			config := Config{
				RateLimitPerSecond: 2,
				RateLimitBurst:     1,
				StatsInstance:      mockStats,
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := config.RateLimit(handler)

			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr

			rateLimited := false
			for i := 0; i < tt.requests; i++ {
				rr := httptest.NewRecorder()
				middleware.ServeHTTP(rr, req)
				if rr.Code == http.StatusTooManyRequests {
					rateLimited = true
					break
				}
			}

			if rateLimited != tt.expectRateLimit {
				t.Errorf("Rate limiting behavior incorrect. Expected rate limiting: %v, got: %v", tt.expectRateLimit, rateLimited)
			}
		})
	}
}
