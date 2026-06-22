package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// requestLog is the structured JSON payload written for every request.
type requestLog struct {
	Level     string `json:"level"`
	Timestamp string `json:"timestamp"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	DurationMs float64 `json:"duration_ms"`
	RequestID string `json:"request_id"`
	RemoteAddr string `json:"remote_addr"`
}

// StructuredLogger returns middleware that emits one JSON log line per
// request to stdout. Requests to /health are skipped to avoid noise from
// Kubernetes liveness probes.
func StructuredLogger() func(http.Handler) http.Handler {
	logger := log.New(os.Stdout, "", 0) // No prefix; we control the format.

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip /health to keep probe traffic out of logs.
			if strings.HasPrefix(r.URL.Path, "/health") {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			// Wrap the ResponseWriter so we can capture the status code.
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			entry := requestLog{
				Level:      "info",
				Timestamp:  start.UTC().Format(time.RFC3339Nano),
				Method:     r.Method,
				Path:       r.URL.Path,
				Status:     ww.Status(),
				DurationMs: float64(time.Since(start).Microseconds()) / 1000.0,
				RequestID:  chimw.GetReqID(r.Context()),
				RemoteAddr: r.RemoteAddr,
			}

			data, err := json.Marshal(entry)
			if err != nil {
				// Extremely unlikely; fall back to unstructured logging.
				logger.Printf("failed to marshal request log: %v", err)
				return
			}
			logger.Println(string(data))
		})
	}
}
