package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/your-username/log-platform/models"
	"github.com/your-username/log-platform/repository"
)

// Trace handles GET /api/v1/trace/{traceId}.
//
// It retrieves all spans for the given trace, computes the unique set of
// services involved, and calculates the total trace duration.
func Trace(repo *repository.ClickHouseRepo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := chi.URLParam(r, "traceId")
		if traceID == "" {
			writeError(w, http.StatusBadRequest, "missing traceId", "traceId path parameter is required")
			return
		}

		start := time.Now()
		spans, err := repo.GetTrace(r.Context(), traceID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "trace query failed", err.Error())
			return
		}

		if len(spans) == 0 {
			writeError(w, http.StatusNotFound, "trace not found",
				"no spans found for trace_id: "+traceID)
			return
		}

		// Collect unique services.
		serviceSet := make(map[string]struct{})
		for _, s := range spans {
			if s.Service != "" {
				serviceSet[s.Service] = struct{}{}
			}
		}
		services := make([]string, 0, len(serviceSet))
		for svc := range serviceSet {
			services = append(services, svc)
		}

		// Compute overall trace duration from the first to the last event.
		durationMs := computeTraceDuration(spans)

		writeJSON(w, http.StatusOK, models.TraceResponse{
			TraceID:    traceID,
			Spans:      spans,
			Services:   services,
			DurationMs: durationMs,
		})
		_ = start // used implicitly by writeJSON timing if extended later.
	}
}

// computeTraceDuration returns the wall-clock duration of a trace in
// milliseconds, measured as the time between the earliest and latest span
// timestamps.
func computeTraceDuration(spans []models.TraceSpan) float64 {
	if len(spans) == 0 {
		return 0
	}
	earliest := spans[0].Timestamp
	latest := spans[0].Timestamp

	for _, s := range spans[1:] {
		if s.Timestamp.Before(earliest) {
			earliest = s.Timestamp
		}
		if s.Timestamp.After(latest) {
			latest = s.Timestamp
		}
	}

	return float64(latest.Sub(earliest).Milliseconds())
}
