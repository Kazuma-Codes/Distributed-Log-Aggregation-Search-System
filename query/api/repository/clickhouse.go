// Package repository provides the ClickHouse data-access layer for the log
// query API. All queries use parameterized inputs to prevent injection.
package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/your-username/log-platform/models"
)

// ClickHouseRepo wraps a ClickHouse connection pool and exposes methods for
// querying the logs table.
type ClickHouseRepo struct {
	conn driver.Conn
}

// New opens a ClickHouse connection pool and returns a ready-to-use repository.
func New(addr, database, username, password string) (*ClickHouseRepo, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:     10 * time.Second,
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 10 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("clickhouse open: %w", err)
	}

	return &ClickHouseRepo{conn: conn}, nil
}

// Ping verifies that the ClickHouse server is reachable and responsive.
func (r *ClickHouseRepo) Ping(ctx context.Context) error {
	return r.conn.Ping(ctx)
}

// Close gracefully shuts down the connection pool.
func (r *ClickHouseRepo) Close() error {
	return r.conn.Close()
}

// SearchLogs queries the logs table with the supplied filter parameters.
// It returns the matching rows and the total count for pagination.
func (r *ClickHouseRepo) SearchLogs(ctx context.Context, p models.SearchParams) ([]models.LogEntry, uint64, error) {
	// --- Build WHERE clause dynamically ---
	var (
		conditions []string
		namedArgs  []driver.NamedValue
	)

	if p.Query != "" {
		conditions = append(conditions, "message ILIKE @query")
		namedArgs = append(namedArgs, driver.NamedValue{Name: "query", Value: "%" + p.Query + "%"})
	}
	if p.Service != "" {
		conditions = append(conditions, "service = @service")
		namedArgs = append(namedArgs, driver.NamedValue{Name: "service", Value: p.Service})
	}
	if p.Level != "" {
		conditions = append(conditions, "level = @level")
		namedArgs = append(namedArgs, driver.NamedValue{Name: "level", Value: p.Level})
	}
	if !p.From.IsZero() {
		conditions = append(conditions, "timestamp >= @from_ts")
		namedArgs = append(namedArgs, driver.NamedValue{Name: "from_ts", Value: p.From})
	}
	if !p.To.IsZero() {
		conditions = append(conditions, "timestamp <= @to_ts")
		namedArgs = append(namedArgs, driver.NamedValue{Name: "to_ts", Value: p.To})
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// --- Count query ---
	countQuery := fmt.Sprintf("SELECT count() FROM logs %s", where)
	var total uint64

	countRow := r.conn.QueryRow(ctx, countQuery, namedArgsToAny(namedArgs)...)
	if err := countRow.Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	// --- Data query ---
	dataQuery := fmt.Sprintf(
		`SELECT timestamp, service, level, message, http_method, http_path,
		        http_status, duration_ms, trace_id, span_id, host,
		        environment, user_id, error_code, stack_trace
		 FROM logs %s
		 ORDER BY timestamp DESC
		 LIMIT @limit OFFSET @offset`,
		where,
	)
	dataArgs := append(namedArgs,
		driver.NamedValue{Name: "limit", Value: p.Limit},
		driver.NamedValue{Name: "offset", Value: p.Offset},
	)

	rows, err := r.conn.Query(ctx, dataQuery, namedArgsToAny(dataArgs)...)
	if err != nil {
		return nil, 0, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var results []models.LogEntry
	for rows.Next() {
		var entry models.LogEntry
		if err := rows.Scan(
			&entry.Timestamp, &entry.Service, &entry.Level, &entry.Message,
			&entry.HTTPMethod, &entry.HTTPPath, &entry.HTTPStatus,
			&entry.DurationMs, &entry.TraceID, &entry.SpanID, &entry.Host,
			&entry.Environment, &entry.UserID, &entry.ErrorCode,
			&entry.StackTrace,
		); err != nil {
			return nil, 0, fmt.Errorf("scan row: %w", err)
		}
		results = append(results, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration: %w", err)
	}

	return results, total, nil
}

// GetTrace retrieves all log entries (spans) belonging to the given trace ID,
// ordered chronologically.
func (r *ClickHouseRepo) GetTrace(ctx context.Context, traceID string) ([]models.TraceSpan, error) {
	const query = `
		SELECT timestamp, service, level, message, http_method, http_path,
		       http_status, duration_ms, trace_id, span_id, host,
		       environment, user_id, error_code, stack_trace,
		       parent_span_id
		FROM logs
		WHERE trace_id = @traceId
		ORDER BY timestamp ASC`

	rows, err := r.conn.Query(ctx, query, clickhouse.Named("traceId", traceID))
	if err != nil {
		return nil, fmt.Errorf("trace query: %w", err)
	}
	defer rows.Close()

	var spans []models.TraceSpan
	for rows.Next() {
		var s models.TraceSpan
		if err := rows.Scan(
			&s.Timestamp, &s.Service, &s.Level, &s.Message,
			&s.HTTPMethod, &s.HTTPPath, &s.HTTPStatus,
			&s.DurationMs, &s.TraceID, &s.SpanID, &s.Host,
			&s.Environment, &s.UserID, &s.ErrorCode,
			&s.StackTrace, &s.ParentSpanID,
		); err != nil {
			return nil, fmt.Errorf("scan span: %w", err)
		}
		spans = append(spans, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return spans, nil
}

// GetStats returns aggregate statistics (error rate, log volume, latency
// percentiles) for the specified time range and optional service filter.
func (r *ClickHouseRepo) GetStats(ctx context.Context, from, to time.Time, service string) (*models.StatsResponse, error) {
	// --- Build WHERE clause ---
	conditions := []string{"timestamp >= @from_ts", "timestamp <= @to_ts"}
	namedArgs := []driver.NamedValue{
		{Name: "from_ts", Value: from},
		{Name: "to_ts", Value: to},
	}
	if service != "" {
		conditions = append(conditions, "service = @service")
		namedArgs = append(namedArgs, driver.NamedValue{Name: "service", Value: service})
	}
	where := "WHERE " + strings.Join(conditions, " AND ")

	// --- Summary: total, error rate, percentiles ---
	summaryQuery := fmt.Sprintf(`
		SELECT
			count()                                          AS total,
			countIf(level = 'ERROR') / greatest(count(), 1)  AS error_rate,
			quantile(0.50)(duration_ms) AS p50,
			quantile(0.90)(duration_ms) AS p90,
			quantile(0.95)(duration_ms) AS p95,
			quantile(0.99)(duration_ms) AS p99
		FROM logs %s`, where)

	var stats models.StatsResponse
	row := r.conn.QueryRow(ctx, summaryQuery, namedArgsToAny(namedArgs)...)
	if err := row.Scan(
		&stats.TotalLogs, &stats.ErrorRate,
		&stats.Percentiles.P50, &stats.Percentiles.P90,
		&stats.Percentiles.P95, &stats.Percentiles.P99,
	); err != nil {
		return nil, fmt.Errorf("stats summary: %w", err)
	}

	// --- Timeseries: log volume bucketed by hour ---
	volumeQuery := fmt.Sprintf(`
		SELECT
			toStartOfHour(timestamp) AS bucket,
			count()                  AS cnt
		FROM logs %s
		GROUP BY bucket
		ORDER BY bucket`, where)

	rows, err := r.conn.Query(ctx, volumeQuery, namedArgsToAny(namedArgs)...)
	if err != nil {
		return nil, fmt.Errorf("volume query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pt models.TimeseriesPt
		if err := rows.Scan(&pt.Bucket, &pt.Count); err != nil {
			return nil, fmt.Errorf("scan volume: %w", err)
		}
		stats.LogVolume = append(stats.LogVolume, pt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("volume rows: %w", err)
	}

	return &stats, nil
}

// GetServices returns a list of distinct services with their log counts.
func (r *ClickHouseRepo) GetServices(ctx context.Context) ([]models.ServiceStats, error) {
	const query = `
		SELECT service, count() AS log_count
		FROM logs
		GROUP BY service
		ORDER BY log_count DESC`

	rows, err := r.conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("services query: %w", err)
	}
	defer rows.Close()

	var services []models.ServiceStats
	for rows.Next() {
		var s models.ServiceStats
		if err := rows.Scan(&s.Service, &s.LogCount); err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}
		services = append(services, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("services rows: %w", err)
	}

	return services, nil
}

// namedArgsToAny converts a slice of driver.NamedValue to []any so it can
// be spread into Query/QueryRow variadic parameters.
func namedArgsToAny(args []driver.NamedValue) []any {
	out := make([]any, len(args))
	for i, a := range args {
		out[i] = clickhouse.Named(a.Name, a.Value)
	}
	return out
}
