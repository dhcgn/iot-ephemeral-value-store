package middleware

import (
	"net"
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

func mustParseCIDR(s string) *net.IPNet {
	_, network, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return network
}

func TestRealIP(t *testing.T) {
	trustedNet := mustParseCIDR("172.19.0.0/16")

	tests := []struct {
		name           string
		remoteAddr     string
		xRealIP        string
		xForwardedFor  string
		trustedProxies []*net.IPNet
		want           string
	}{
		{
			name:       "Uses RemoteAddr when no proxy headers",
			remoteAddr: "1.2.3.4:5678",
			want:       "1.2.3.4",
		},
		{
			name:           "Prefers X-Real-IP over RemoteAddr when proxy is trusted",
			remoteAddr:     "172.19.0.7:1234",
			xRealIP:        "203.0.113.5",
			trustedProxies: []*net.IPNet{trustedNet},
			want:           "203.0.113.5",
		},
		{
			name:           "Uses X-Forwarded-For when X-Real-IP absent and proxy is trusted",
			remoteAddr:     "172.19.0.7:1234",
			xForwardedFor:  "203.0.113.10, 10.0.0.1",
			trustedProxies: []*net.IPNet{trustedNet},
			want:           "203.0.113.10",
		},
		{
			name:           "Prefers X-Real-IP over X-Forwarded-For",
			remoteAddr:     "172.19.0.7:1234",
			xRealIP:        "203.0.113.5",
			xForwardedFor:  "203.0.113.10",
			trustedProxies: []*net.IPNet{trustedNet},
			want:           "203.0.113.5",
		},
		{
			name:           "X-Forwarded-For with single entry",
			remoteAddr:     "172.19.0.7:1234",
			xForwardedFor:  "198.51.100.42",
			trustedProxies: []*net.IPNet{trustedNet},
			want:           "198.51.100.42",
		},
		{
			name:           "Ignores proxy headers when proxy IP is not trusted",
			remoteAddr:     "10.0.0.1:1234",
			xRealIP:        "203.0.113.5",
			xForwardedFor:  "203.0.113.10",
			trustedProxies: []*net.IPNet{trustedNet},
			want:           "10.0.0.1",
		},
		{
			name:          "Ignores proxy headers when TrustedProxies is empty",
			remoteAddr:    "172.19.0.7:1234",
			xRealIP:       "203.0.113.5",
			xForwardedFor: "203.0.113.10",
			want:          "172.19.0.7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}

			got := realIP(req, tt.trustedProxies)
			if got != tt.want {
				t.Errorf("realIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRateLimit_BehindReverseProxy(t *testing.T) {
	// Simulate Traefik forwarding requests from two different real clients.
	// Both arrive with the same RemoteAddr (Traefik's internal IP), but with
	// distinct X-Forwarded-For values. Each client should have its own rate
	// limit bucket.
	trustedNet := mustParseCIDR("172.19.0.0/16")
	mockStats := stats.NewStats()
	config := Config{
		RateLimitPerSecond: 2,
		RateLimitBurst:     1,
		StatsInstance:      mockStats,
		TrustedProxies:     []*net.IPNet{trustedNet},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := config.RateLimit(handler)

	makeReq := func(xff string) *http.Request {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "172.19.0.7:12345"
		req.Header.Set("X-Forwarded-For", xff)
		return req
	}

	// Exhaust the bucket for client A.
	reqA := makeReq("203.0.113.1")
	for i := 0; i < 10; i++ {
		rr := httptest.NewRecorder()
		mw.ServeHTTP(rr, reqA)
		if rr.Code == http.StatusTooManyRequests {
			break
		}
	}

	// Client B should still be allowed on its first request.
	reqB := makeReq("203.0.113.2")
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, reqB)
	if rr.Code == http.StatusTooManyRequests {
		t.Error("client B should not be rate limited when only client A has exceeded the limit")
	}
}
