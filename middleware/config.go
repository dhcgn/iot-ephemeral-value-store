package middleware

import (
	"net"

	"github.com/dhcgn/iot-ephemeral-value-store/stats"
)

// Config holds the configuration for all middleware functions.
type Config struct {
	MaxRequestSize     int64
	RateLimitPerSecond float64
	RateLimitBurst     int
	StatsInstance      *stats.Stats

	// TrustedProxies is a list of IP networks whose requests are allowed to
	// override the client IP via X-Real-IP or X-Forwarded-For headers. When
	// empty, proxy headers are ignored and r.RemoteAddr is used directly.
	// Set this to the CIDR of your reverse proxy (e.g. Traefik) Docker network.
	TrustedProxies []*net.IPNet
}
