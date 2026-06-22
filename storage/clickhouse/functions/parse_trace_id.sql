-- Custom function to parse and validate trace IDs.
-- Ensures that the trace_id conforms to expected formats (e.g., 32 hex characters for W3C trace context).

CREATE FUNCTION IF NOT EXISTS parseTraceId AS (raw) -> 
    multiIf(
        -- If length is exactly 32 and it contains only hex chars, assume it's valid W3C trace_id
        match(raw, '^[a-fA-F0-9]{32}$'), lower(raw),
        -- Fallback: extract continuous hex string of 32 chars from raw string if embedded
        match(raw, '[a-fA-F0-9]{32}'), lower(extract(raw, '[a-fA-F0-9]{32}')),
        -- Default to empty if no valid trace ID found
        ''
    );
