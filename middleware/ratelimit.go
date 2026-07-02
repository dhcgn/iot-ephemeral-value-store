package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

// realIP extracts the real client IP from the request. When the direct peer
// address matches one of the trusted proxy networks, the actual client address
// is read from X-Real-IP or X-Forwarded-For headers set by the proxy. Falls
// back to r.RemoteAddr when no trusted proxy headers are present.
func realIP(r *http.Request, trustedProxies []*net.IPNet) string {
	peerIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		peerIP = r.RemoteAddr
	}

	if len(trustedProxies) > 0 && isTrustedProxy(peerIP, trustedProxies) {
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// X-Forwarded-For may be a comma-separated list; the first entry is
			// the original client IP.
			return strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
		}
	}

	return peerIP
}

// isTrustedProxy reports whether ip is contained in any of the given networks.
func isTrustedProxy(ip string, networks []*net.IPNet) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, network := range networks {
		if network.Contains(parsed) {
			return true
		}
	}
	return false
}

// RateLimit is a middleware that limits the number of requests per second per
// client IP. Local addresses are excluded from rate limiting. When
// TrustedProxies is configured, the real client IP is resolved from proxy
// headers so that each end-user IP gets its own rate limit bucket even when all
// traffic arrives through a single reverse proxy such as Traefik.
func (c Config) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In local testing, r.RemoteAddr is empty
		if r.RemoteAddr == "" {
			next.ServeHTTP(w, r)
			return
		}

		ip := realIP(r, c.TrustedProxies)
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
