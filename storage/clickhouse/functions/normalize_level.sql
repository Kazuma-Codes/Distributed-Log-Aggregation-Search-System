-- Normalize log levels from various formats to standard taxonomy.
-- Map: FATAL/CRITICAL->critical, ERR/ERROR->error, WARN/WARNING->warn, INFO/INFORMATION->info, DBG/DEBUG/TRACE->debug

CREATE FUNCTION IF NOT EXISTS normalizeLevel AS (raw) -> 
    multiIf(
        upper(raw) IN ('FATAL', 'CRITICAL', 'EMERGENCY'), 'critical',
        upper(raw) IN ('ERR', 'ERROR'), 'error',
        upper(raw) IN ('WARN', 'WARNING'), 'warn',
        upper(raw) IN ('INFO', 'INFORMATION'), 'info',
        upper(raw) IN ('DBG', 'DEBUG', 'TRACE'), 'debug',
        'unknown'
    );
