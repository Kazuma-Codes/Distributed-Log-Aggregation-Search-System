// Package handlers implements the HTTP handler functions for the log query API.
// Each handler parses request parameters, delegates to the repository layer,
// and returns a JSON response.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/your-username/log-platform/models"
	"github.com/your-username/log-platform/repository"
)

// Search handles GET /api/v1/search.
//
// Query parameters:
//   - q:       free-text search (ILIKE)
//   - service: exact match on service name
//   - level:   exact match on log level (e.g. ERROR, WARN)
//   - from:    start of time range (RFC3339)
//   - to:      end of time range (RFC3339)
//   - limit:   max results per page (default 100, max 1000)
//   - offset:  pagination offset (default 0)
func Search(repo *repository.ClickHouseRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		params := models.SearchParams{
			Query:   q.Get("q"),
			Service: q.Get("service"),
			Level:   q.Get("level"),
		}

		// Parse time range.
		if v := q.Get("from"); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid 'from' timestamp", err.Error())
				return
			}
			params.From = t
		}
		if v := q.Get("to"); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid 'to' timestamp", err.Error())
				return
			}
			params.To = t
		}

		// Parse pagination.
		params.Limit = intQueryParam(q.Get("limit"), 100)
		if params.Limit <= 0 {
			params.Limit = 100
		}
		if params.Limit > 1000 {
			params.Limit = 1000
		}

		params.Offset = intQueryParam(q.Get("offset"), 0)
		if params.Offset < 0 {
			params.Offset = 0
		}

		start := time.Now()
		entries, total, err := repo.SearchLogs(r.Context(), params)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "search failed", err.Error())
			return
		}

		// Ensure we return an empty array instead of null.
		if entries == nil {
			entries = []models.LogEntry{}
		}

		writeJSON(w, http.StatusOK, models.SearchResponse{
			Data:        entries,
			Total:       total,
			Limit:       params.Limit,
			Offset:      params.Offset,
			QueryTimeMs: time.Since(start).Milliseconds(),
		})
	}
}

// intQueryParam parses s as an integer, returning fallback on failure.
func intQueryParam(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}

// writeJSON marshals v to JSON and writes it to the response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a standardised JSON error response.
func writeError(w http.ResponseWriter, status int, msg, details string) {
	writeJSON(w, status, models.ErrorResponse{
		Error:   msg,
		Code:    status,
		Details: details,
	})
}
