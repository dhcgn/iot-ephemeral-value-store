package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

// realIP extracts the real client IP from the request. When the server is
// deployed behind a reverse proxy (e.g. Traefik), the actual client address is
// forwarded via the X-Real-IP or X-Forwarded-For headers. Falls back to
// r.RemoteAddr when no proxy headers are present.
func realIP(r *http.Request) string {
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For may be a comma-separated list; the first entry is the
		// original client IP.
		return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// RateLimit is a middleware that limits the number of requests per second per
// client IP. Local addresses and requests forwarded by a trusted reverse proxy
// that originate from localhost are excluded from rate limiting.
func (c Config) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In local testing, r.RemoteAddr is empty
		if r.RemoteAddr == "" {
			next.ServeHTTP(w, r)
			return
		}

		ip := realIP(r)
		if ip == "" {
			slog.Error("middleware: failed to parse remote address", "remote_addr", r.RemoteAddr)
			c.StatsInstance.IncrementHTTPErrors()
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Exclude local IP addresses from rate limiting
		if ip == "127.0.0.1" || ip == "::1" {
			next.ServeHTTP(w, r)
			return
		}

		limiter := c.getLimiter(ip)
		if !limiter.Allow() {
			slog.Error("middleware: rate limit exceeded", "remote_addr", ip, "method", r.Method, "path", r.URL.Path)
			c.StatsInstance.IncrementHTTPErrors()
			c.StatsInstance.RecordRateLimitHit(ip)
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (c Config) getLimiter(ip string) *rate.Limiter {
	mtx.Lock()
	defer mtx.Unlock()

	limiter, exists := clients[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(c.RateLimitPerSecond), c.RateLimitBurst)
		clients[ip] = limiter
	}

	return limiter
}

var clients = make(map[string]*rate.Limiter)
var mtx sync.Mutex
