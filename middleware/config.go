package middleware

import "github.com/dhcgn/iot-ephemeral-value-store/stats"

type Config struct {
	MaxRequestSize     int64
	RateLimitPerSecond float64
	RateLimitBurst     int
	StatsInstance      *stats.Stats
}
