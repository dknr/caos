package server

import (
	"net/http"
)

// APIKeyAuth returns a middleware that requires a valid API key (X-API-Key header)
// for write operations (POST, PUT, DELETE). Read operations (GET) are always allowed.
// When no API key is configured, write operations are blocked entirely.
func APIKeyAuth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read operations are always allowed
			if r.Method == http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Write operations require a valid API key
			if r.Header.Get("X-API-Key") != apiKey {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
