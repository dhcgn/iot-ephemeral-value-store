package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// rateLimit is a middleware that limits the number of requests per second
// Exclude local IP addresses from rate limiting for debugging and testing purposes
func (c Config) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In locla testing, r.RemoteAddr is empty
		if r.RemoteAddr == "" {
			next.ServeHTTP(w, r)
			return
		}

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			slog.Error("middleware: failed to parse remote address", "error", err, "remote_addr", r.RemoteAddr)
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
