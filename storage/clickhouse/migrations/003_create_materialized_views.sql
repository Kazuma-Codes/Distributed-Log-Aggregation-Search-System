-- Materialized Views to route data from the Kafka engine to persistent MergeTree tables.
-- Design decisions:
-- 1. Main View (logs_mv): Simply reads from the Kafka queue and inserts into the main `logs` table. This is the primary ingestion path.
-- 2. Error Logs (logs_errors): A separate table strictly for error and critical logs. Separating them allows for faster alerting and querying since error volume is typically much smaller than total volume.
-- 3. Error View (logs_errors_mv): Filters the stream on the fly and routes only errors to the `logs_errors` table.

-- Route all logs to the main logs table
CREATE MATERIALIZED VIEW IF NOT EXISTS logs_mv TO logs AS
SELECT * FROM logs_queue;

-- Create a dedicated table for errors for fast access
CREATE TABLE IF NOT EXISTS logs_errors (
    timestamp DateTime64(3),
    service LowCardinality(String),
    level LowCardinality(String),
    message String,
    http_method LowCardinality(String) DEFAULT '',
    http_path String DEFAULT '',
    http_status UInt16 DEFAULT 0,
    duration_ms Float64 DEFAULT 0,
    trace_id String DEFAULT '',
    span_id String DEFAULT '',
    parent_span_id String DEFAULT '',
    host LowCardinality(String) DEFAULT '',
    environment LowCardinality(String) DEFAULT 'production',
    datacenter LowCardinality(String) DEFAULT 'dc1',
    user_id String DEFAULT '',
    correlation_id String DEFAULT '',
    request_size UInt32 DEFAULT 0,
    response_size UInt32 DEFAULT 0,
    error_code String DEFAULT '',
    stack_trace String DEFAULT '',
    
    INDEX idx_message message TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 4,
    INDEX idx_trace_id trace_id TYPE bloom_filter(0.01) GRANULARITY 1
) ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (service, timestamp, trace_id)
TTL timestamp + INTERVAL 7 DAY TO VOLUME 'warm',
    timestamp + INTERVAL 90 DAY TO VOLUME 'cold';

-- Route only error logs to the dedicated errors table
CREATE MATERIALIZED VIEW IF NOT EXISTS logs_errors_mv TO logs_errors AS
SELECT * FROM logs_queue WHERE level IN ('error', 'critical', 'fatal');
