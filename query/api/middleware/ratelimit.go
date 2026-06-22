package middleware

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/your-username/log-platform/models"
)

// ipLimiter pairs a rate.Limiter with the last time it was used,
// so stale entries can be cleaned up.
type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter returns middleware that enforces per-IP rate limiting.
// rps is the sustained requests-per-second and burst is the token-bucket
// burst size. A background goroutine evicts entries that have been idle
// for more than 3 minutes.
func RateLimiter(rps float64, burst int) func(http.Handler) http.Handler {
	var (
		mu       sync.Mutex
		limiters = make(map[string]*ipLimiter)
	)

	// Cleanup goroutine: evict IP entries that have been idle for 3 minutes.
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for ip, l := range limiters {
				if time.Since(l.lastSeen) > 3*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()

	// getLimiter retrieves or creates a limiter for the given IP.
	getLimiter := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()

		entry, ok := limiters[ip]
		if !ok {
			entry = &ipLimiter{
				limiter: rate.NewLimiter(rate.Limit(rps), burst),
			}
			limiters[ip] = entry
		}
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the client IP, stripping any port number.
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr // Fallback: use raw value.
			}

			limiter := getLimiter(ip)
			if !limiter.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(models.ErrorResponse{
					Error:   "rate limit exceeded",
					Code:    http.StatusTooManyRequests,
					Details: "too many requests — slow down and retry after the Retry-After period",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
