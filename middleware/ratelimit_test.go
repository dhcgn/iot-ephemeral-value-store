package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/stats"
)

func TestRateLimit(t *testing.T) {
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

	// Test non-local IP (should be rate limited)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	var responses []int
	for i := 0; i < 10; i++ {
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
		responses = append(responses, rr.Code)
	}

	// Check if any rate limiting occurred
	rateLimited := false
	for _, status := range responses {
		if status == http.StatusTooManyRequests {
			rateLimited = true
			break
		}
	}

	if !rateLimited {
		t.Errorf("Rate limiting did not occur for non-local IP")
	}

	// Test local IP (should not be rate limited)
	req = httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"

	for i := 0; i < 10; i++ {
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("Local IP was rate limited: got status %d", rr.Code)
		}
	}
}
