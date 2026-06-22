-- Pre-aggregated tables for dashboard performance
-- Design decisions:
-- 1. SummingMergeTree and AggregatingMergeTree: Used to aggregate data on insertion or via background merges.
-- 2. 5-minute / 1-minute buckets: Optimized intervals for common dashboard charting and anomaly detection.
-- 3. Materialized Views: Used to populate the aggregation tables seamlessly from the main `logs` stream (or `logs_queue`).

-- 1. Error rate by service (5-minute buckets)
CREATE TABLE IF NOT EXISTS agg_error_rate_5m (
    bucket DateTime,
    service LowCardinality(String),
    level LowCardinality(String),
    count UInt64,
    error_count UInt64
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMMDD(bucket)
ORDER BY (service, level, bucket);

CREATE MATERIALIZED VIEW IF NOT EXISTS agg_error_rate_5m_mv TO agg_error_rate_5m AS
SELECT
    toStartOfFiveMinutes(timestamp) as bucket,
    service,
    level,
    count() as count,
    countIf(level IN ('error', 'critical', 'fatal')) as error_count
FROM logs
GROUP BY bucket, service, level;

-- 2. Latency percentiles by service (5-minute buckets)
-- Using AggregateFunction and AggregatingMergeTree for quantiles
CREATE TABLE IF NOT EXISTS agg_latency_5m (
    bucket DateTime,
    service LowCardinality(String),
    http_path String,
    request_count UInt64,
    duration_quantiles AggregateFunction(quantiles(0.5, 0.9, 0.95, 0.99), Float64)
) ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMMDD(bucket)
ORDER BY (service, http_path, bucket);

CREATE MATERIALIZED VIEW IF NOT EXISTS agg_latency_5m_mv TO agg_latency_5m AS
SELECT
    toStartOfFiveMinutes(timestamp) as bucket,
    service,
    http_path,
    count() as request_count,
    quantilesState(0.5, 0.9, 0.95, 0.99)(duration_ms) as duration_quantiles
FROM logs
WHERE duration_ms > 0
GROUP BY bucket, service, http_path;

-- 3. Log volume by service (1-minute buckets) for anomaly detection
CREATE TABLE IF NOT EXISTS agg_log_volume_1m (
    bucket DateTime,
    service LowCardinality(String),
    count UInt64
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMMDD(bucket)
ORDER BY (service, bucket);

CREATE MATERIALIZED VIEW IF NOT EXISTS agg_log_volume_1m_mv TO agg_log_volume_1m AS
SELECT
    toStartOfMinute(timestamp) as bucket,
    service,
    count() as count
FROM logs
GROUP BY bucket, service;

-- 4. HTTP status distribution (5-minute buckets)
CREATE TABLE IF NOT EXISTS agg_http_status_5m (
    bucket DateTime,
    service LowCardinality(String),
    http_status UInt16,
    count UInt64
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMMDD(bucket)
ORDER BY (service, http_status, bucket);

CREATE MATERIALIZED VIEW IF NOT EXISTS agg_http_status_5m_mv TO agg_http_status_5m AS
SELECT
    toStartOfFiveMinutes(timestamp) as bucket,
    service,
    http_status,
    count() as count
FROM logs
WHERE http_status > 0
GROUP BY bucket, service, http_status;
