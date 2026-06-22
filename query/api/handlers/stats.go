package handlers

import (
	"net/http"
	"time"

	"github.com/your-username/log-platform/repository"
)

// Stats handles GET /api/v1/stats.
//
// Query parameters:
//   - from:    start of time range (RFC3339, required)
//   - to:      end of time range (RFC3339, required)
//   - service: optional service filter
//
// Returns error rates, log volume, and latency percentiles.
func Stats(repo *repository.ClickHouseRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		from, err := time.Parse(time.RFC3339, q.Get("from"))
		if err != nil {
			// Default to last 24 hours if not specified.
			from = time.Now().Add(-24 * time.Hour)
		}

		to, err := time.Parse(time.RFC3339, q.Get("to"))
		if err != nil {
			to = time.Now()
		}

		service := q.Get("service")

		start := time.Now()
		stats, err := repo.GetStats(r.Context(), from, to, service)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "stats query failed", err.Error())
			return
		}

		stats.QueryTimeMs = time.Since(start).Milliseconds()
		writeJSON(w, http.StatusOK, stats)
	}
}

// Services handles GET /api/v1/stats/services.
//
// Returns a list of distinct services with their log counts.
func Services(repo *repository.ClickHouseRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		services, err := repo.GetServices(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "services query failed", err.Error())
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"services":      services,
			"query_time_ms": time.Since(start).Milliseconds(),
		})
	}
}
