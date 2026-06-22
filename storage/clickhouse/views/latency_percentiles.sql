-- View for latency analysis over the past hour.
-- Calculates percentiles (p50, p90, p95, p99) and averages to identify performance bottlenecks.

CREATE VIEW IF NOT EXISTS v_latency_percentiles AS
SELECT
    service,
    http_path,
    count() as request_count,
    round(quantile(0.50)(duration_ms), 2) as p50,
    round(quantile(0.90)(duration_ms), 2) as p90,
    round(quantile(0.95)(duration_ms), 2) as p95,
    round(quantile(0.99)(duration_ms), 2) as p99,
    round(max(duration_ms), 2) as max_ms,
    round(avg(duration_ms), 2) as avg_ms
FROM logs
WHERE timestamp > now() - INTERVAL 1 HOUR
    AND duration_ms > 0
GROUP BY service, http_path
ORDER BY p99 DESC;
