package middleware

import (
	"net/http"
)

func (c Config) LimitRequestSize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request size is too large
		if r.ContentLength > c.MaxRequestSize {
			c.StatsInstance.IncrementHTTPErrors()
			http.Error(w, "Request size is too large", http.StatusRequestEntityTooLarge)
			return
		}

		next.ServeHTTP(w, r)
	})
}
