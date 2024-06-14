package middleware

type Config struct {
	MaxRequestSize     int64
	RateLimitPerSecond float64
	RateLimitBurst     int
}
