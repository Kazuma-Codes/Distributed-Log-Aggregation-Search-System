-- View for detecting volume anomalies using a moving average comparison.
-- Design decisions:
-- 1. Uses Common Table Expressions (WITH) to calculate the current volume and the historical volume.
-- 2. Compares the current 5-min bucket to the 7-day average of the same time window to spot unexpected spikes or drops.

CREATE VIEW IF NOT EXISTS v_log_volume_anomaly AS
WITH current_volume AS (
    SELECT
        service,
        toStartOfFiveMinutes(timestamp) AS current_bucket,
        count() AS current_count
    FROM logs
    WHERE timestamp >= now() - INTERVAL 5 MINUTE
    GROUP BY service, current_bucket
),
historical_volume AS (
    SELECT
        service,
        toStartOfFiveMinutes(timestamp) AS hist_bucket,
        count() AS hist_count
    FROM logs
    WHERE timestamp >= now() - INTERVAL 7 DAY 
      AND timestamp < now() - INTERVAL 5 MINUTE
      -- Only match the time-of-day by filtering for similar hour/minute
      AND toHour(timestamp) = toHour(now())
      AND toMinute(toStartOfFiveMinutes(timestamp)) = toMinute(toStartOfFiveMinutes(now()))
    GROUP BY service, hist_bucket
),
avg_historical AS (
    SELECT
        service,
        avg(hist_count) AS avg_past_count,
        stddevPop(hist_count) AS stddev_past_count
    FROM historical_volume
    GROUP BY service
)
SELECT
    c.service,
    c.current_bucket,
    c.current_count,
    round(h.avg_past_count, 2) AS expected_count,
    round(h.stddev_past_count, 2) AS stddev_count,
    -- Simple z-score to find anomaly severity
    multiIf(
        h.stddev_past_count = 0, 0,
        round(abs(c.current_count - h.avg_past_count) / h.stddev_past_count, 2)
    ) AS anomaly_score,
    (c.current_count > h.avg_past_count + 3 * h.stddev_past_count) OR 
    (c.current_count < h.avg_past_count - 3 * h.stddev_past_count AND h.avg_past_count > 100) AS is_anomaly
FROM current_volume c
LEFT JOIN avg_historical h ON c.service = h.service
ORDER BY anomaly_score DESC;
