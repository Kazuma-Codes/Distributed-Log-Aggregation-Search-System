-- Main logs table using MergeTree engine for efficient storage and querying of log data.
-- Design decisions:
-- 1. Partitioning by Day (toYYYYMMDD(timestamp)): Allows dropping old data efficiently and limits the amount of data to scan for time-bounded queries.
-- 2. Sorting (ORDER BY): Ordered by service, timestamp, trace_id to optimize filtering by service and time, and looking up specific traces.
-- 3. LowCardinality: Used for low-variance fields (service, level, host, environment, datacenter) to compress data and speed up string comparisons.
-- 4. TTL: Data automatically moves to 'warm' storage after 7 days, and to 'cold' storage after 90 days.
-- 5. Bloom Filters: Used on message (tokenbf_v1), trace_id, user_id, and error_code to quickly skip blocks that don't contain the searched terms.
-- 6. MinMax Index: Used on http_status for fast filtering of status code ranges.

CREATE TABLE IF NOT EXISTS logs (
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
    INDEX idx_trace_id trace_id TYPE bloom_filter(0.01) GRANULARITY 1,
    INDEX idx_user_id user_id TYPE bloom_filter(0.01) GRANULARITY 4,
    INDEX idx_http_status http_status TYPE minmax GRANULARITY 4,
    INDEX idx_error_code error_code TYPE bloom_filter(0.01) GRANULARITY 4
) ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (service, timestamp, trace_id)
TTL timestamp + INTERVAL 7 DAY TO VOLUME 'warm',
    timestamp + INTERVAL 90 DAY TO VOLUME 'cold'
SETTINGS
    index_granularity = 8192,
    merge_with_ttl_timeout = 86400;
