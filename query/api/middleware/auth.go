// Package middleware provides HTTP middleware for the log query API,
// including authentication, rate limiting, and structured request logging.
package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/your-username/log-platform/models"
)

// APIKeyAuth returns middleware that validates requests against the provided
// apiKey. If apiKey is empty, authentication is disabled (development mode).
// Health and readiness probes are always exempt from authentication.
func APIKeyAuth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health/readiness probes so K8s can always reach them.
			if strings.HasPrefix(r.URL.Path, "/health") || strings.HasPrefix(r.URL.Path, "/ready") {
				next.ServeHTTP(w, r)
				return
			}

			// Development mode: no API key configured.
			if apiKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Accept the key from the X-API-Key header or the api_key query param.
			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = r.URL.Query().Get("api_key")
			}

			if key != apiKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(models.ErrorResponse{
					Error: "unauthorized",
					Code:  http.StatusUnauthorized,
					Details: "missing or invalid API key — provide X-API-Key header " +
						"or api_key query parameter",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
