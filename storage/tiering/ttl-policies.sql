-- SQL statements to apply/modify TTL policies across the core tables.
-- Design decisions:
-- - Hot -> Warm at 7 days
-- - Warm -> Cold at 90 days
-- - Delete at 365 days (1 year retention)

-- Apply to main logs table
ALTER TABLE logs MODIFY TTL 
    timestamp + INTERVAL 7 DAY TO VOLUME 'warm',
    timestamp + INTERVAL 90 DAY TO VOLUME 'cold',
    timestamp + INTERVAL 365 DAY;

-- Apply to error logs table
ALTER TABLE logs_errors MODIFY TTL 
    timestamp + INTERVAL 7 DAY TO VOLUME 'warm',
    timestamp + INTERVAL 90 DAY TO VOLUME 'cold',
    timestamp + INTERVAL 365 DAY;

-- Apply to aggregated tables (maybe shorter TTL, e.g., keep 5m aggregations for 90 days)
ALTER TABLE agg_error_rate_5m MODIFY TTL 
    bucket + INTERVAL 30 DAY TO VOLUME 'warm',
    bucket + INTERVAL 90 DAY;

ALTER TABLE agg_latency_5m MODIFY TTL 
    bucket + INTERVAL 30 DAY TO VOLUME 'warm',
    bucket + INTERVAL 90 DAY;

ALTER TABLE agg_log_volume_1m MODIFY TTL 
    bucket + INTERVAL 7 DAY TO VOLUME 'warm',
    bucket + INTERVAL 30 DAY;

ALTER TABLE agg_http_status_5m MODIFY TTL 
    bucket + INTERVAL 30 DAY TO VOLUME 'warm',
    bucket + INTERVAL 90 DAY;
