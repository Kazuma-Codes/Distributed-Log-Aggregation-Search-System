package handlers

import (
	"net/http"

	"github.com/your-username/log-platform/repository"
)

// Health handles GET /health.
// It always returns HTTP 200 to satisfy Kubernetes liveness probes.
func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// Ready handles GET /ready.
// It pings ClickHouse and returns HTTP 200 when the database is reachable,
// or HTTP 503 when it is not. This drives Kubernetes readiness probes so
// traffic is only routed to pods that can actually serve queries.
func Ready(repo *repository.ClickHouseRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := repo.Ping(r.Context()); err != nil {
			writeError(w, http.StatusServiceUnavailable, "not ready",
				"clickhouse ping failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}
