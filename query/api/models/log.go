// Package models defines the domain types and API response structures for the
// log query service. These types map directly to the ClickHouse schema and are
// serialized as JSON in HTTP responses.
package models

import "time"

// LogEntry represents a single structured log record stored in ClickHouse.
// Fields tagged with `ch` map to ClickHouse column names; JSON tags control
// the API serialization.
type LogEntry struct {
	Timestamp   time.Time `json:"timestamp" ch:"timestamp"`
	Service     string    `json:"service" ch:"service"`
	Level       string    `json:"level" ch:"level"`
	Message     string    `json:"message" ch:"message"`
	HTTPMethod  string    `json:"http_method,omitempty" ch:"http_method"`
	HTTPPath    string    `json:"http_path,omitempty" ch:"http_path"`
	HTTPStatus  uint16    `json:"http_status,omitempty" ch:"http_status"`
	DurationMs  float64   `json:"duration_ms,omitempty" ch:"duration_ms"`
	TraceID     string    `json:"trace_id,omitempty" ch:"trace_id"`
	SpanID      string    `json:"span_id,omitempty" ch:"span_id"`
	Host        string    `json:"host,omitempty" ch:"host"`
	Environment string    `json:"environment,omitempty" ch:"environment"`
	UserID      string    `json:"user_id,omitempty" ch:"user_id"`
	ErrorCode   string    `json:"error_code,omitempty" ch:"error_code"`
	StackTrace  string    `json:"stack_trace,omitempty" ch:"stack_trace"`
}

// SearchParams holds the parsed and validated query parameters for a log
// search request.
type SearchParams struct {
	Query   string    // Free-text search (ILIKE).
	Service string    // Exact match on service name.
	Level   string    // Exact match on log level.
	From    time.Time // Start of time range (inclusive).
	To      time.Time // End of time range (inclusive).
	Limit   int       // Maximum number of results (capped at 1000).
	Offset  int       // Pagination offset.
}

// SearchResponse is the JSON envelope for search results.
type SearchResponse struct {
	Data        []LogEntry `json:"data"`
	Total       uint64     `json:"total"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
	QueryTimeMs int64      `json:"query_time_ms"`
}

// TraceSpan extends LogEntry with parent information for span-tree
// reconstruction.
type TraceSpan struct {
	LogEntry
	ParentSpanID string `json:"parent_span_id,omitempty" ch:"parent_span_id"`
}

// TraceResponse is the JSON envelope for trace lookup results.
type TraceResponse struct {
	TraceID    string      `json:"trace_id"`
	Spans      []TraceSpan `json:"spans"`
	Services   []string    `json:"services"`
	DurationMs float64     `json:"duration_ms"`
}

// StatsResponse is the JSON envelope for aggregate statistics.
type StatsResponse struct {
	TotalLogs    uint64           `json:"total_logs"`
	ErrorRate    float64          `json:"error_rate"`
	LogVolume    []TimeseriesPt   `json:"log_volume"`
	Percentiles  LatencyPctiles   `json:"latency_percentiles"`
	QueryTimeMs  int64            `json:"query_time_ms"`
}

// TimeseriesPt represents a single point in a time-series aggregation.
type TimeseriesPt struct {
	Bucket time.Time `json:"bucket"`
	Count  uint64    `json:"count"`
}

// LatencyPctiles holds pre-computed latency percentiles.
type LatencyPctiles struct {
	P50 float64 `json:"p50"`
	P90 float64 `json:"p90"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

// ServiceStats represents a service with its log count.
type ServiceStats struct {
	Service  string `json:"service"`
	LogCount uint64 `json:"log_count"`
}

// ServicesResponse is the JSON envelope for the services endpoint.
type ServicesResponse struct {
	Services    []ServiceStats `json:"services"`
	QueryTimeMs int64          `json:"query_time_ms"`
}

// ErrorResponse is the standard JSON error envelope returned by the API.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}
