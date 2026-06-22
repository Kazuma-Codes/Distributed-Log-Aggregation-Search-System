-- View for real-time error rate monitoring.
-- Calculates the percentage of error/critical logs per service over the last 24 hours.

CREATE VIEW IF NOT EXISTS v_error_rate_by_service AS
SELECT
    service,
    toStartOfFiveMinutes(timestamp) as bucket,
    count() as total,
    countIf(level = 'error') as errors,
    countIf(level = 'critical') as criticals,
    round(countIf(level IN ('error', 'critical', 'fatal')) / count() * 100, 2) as error_rate_pct
FROM logs
WHERE timestamp > now() - INTERVAL 24 HOUR
GROUP BY service, bucket
ORDER BY bucket DESC, error_rate_pct DESC;
