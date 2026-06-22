-- GDPR compliance tables
-- Design decisions:
-- 1. Audit Table: Tracks the status of GDPR deletion requests, providing an audit trail.
-- 2. User Sessions Mapping: Enables tracing a user's session and the specific trace IDs associated with their activity, making it possible to identify logs related to a specific user for deletion.

-- Audit table for deletion requests
CREATE TABLE IF NOT EXISTS gdpr_deletion_audit (
    request_id UUID DEFAULT generateUUIDv4(),
    user_id String,
    requested_at DateTime DEFAULT now(),
    completed_at Nullable(DateTime),
    rows_affected UInt64 DEFAULT 0,
    status LowCardinality(String) DEFAULT 'pending',
    requested_by String
) ENGINE = MergeTree()
ORDER BY (requested_at, user_id);

-- User sessions mapping (for trace_id lookup)
CREATE TABLE IF NOT EXISTS user_sessions (
    user_id String,
    trace_id String,
    session_start DateTime,
    session_end Nullable(DateTime),
    service LowCardinality(String)
) ENGINE = MergeTree()
ORDER BY (user_id, session_start);
