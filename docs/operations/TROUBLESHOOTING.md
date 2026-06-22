# Troubleshooting Guide

Common issues and their resolutions.

## 1. Kafka Consumer Lag Increasing
**Symptoms**: Dashboards show log latency increasing. Recent logs are not appearing in searches. Prometheus alerts on high Kafka lag.
**Diagnosis**:
- Check ClickHouse logs: `kubectl logs -l app=clickhouse -c clickhouse`.
- Look for parsing errors or connection timeouts.
- Check ClickHouse CPU usage. If 100%, it cannot ingest fast enough.
**Resolution**:
- If ClickHouse is overloaded, scale ClickHouse vertically (more CPU) or horizontally (more shards).
- If logs contain malformed JSON causing parse errors, fix the application output or add a VRL transform in Vector to sanitize logs before Kafka.
- Temporarily increase the number of Kafka engine consumer threads in ClickHouse.

## 2. ClickHouse Slow Queries
**Symptoms**: API latency > 2 seconds. Grafana dashboards time out.
**Diagnosis**:
- Run `SHOW PROCESSLIST` in ClickHouse shell to identify slow queries.
- Check if the query is scanning too much data (missing date filters).
- Check if the query is hitting the S3 warm/cold tier unexpectedly.
**Resolution**:
- Ensure all queries include a time range filter (`timestamp >= ...`).
- If full-text search is slow, consider adding a skip index (`tokenbf_v1`) to the queried column.
- Warn users about querying older data that resides in the MinIO S3 tier.

## 3. Vector Dropping Logs
**Symptoms**: Applications are logging, but log volume in Grafana drops unexpectedly.
**Diagnosis**:
- Check Vector metrics: `vector_component_discarded_events_total`.
- Look at Vector pod logs for "connection refused" to Kafka.
**Resolution**:
- If Kafka is unreachable, check network policies and Kafka broker status.
- If Vector's disk buffer is full, increase `buffer.max_size` or fix the downstream bottleneck (Kafka).

## 4. Grafana Can't Connect to Datasource
**Symptoms**: Dashboards show "No Data" or "Network Error".
**Diagnosis**:
- Check Grafana pod logs.
- Exec into Grafana pod and `curl` the ClickHouse HTTP port (8123).
**Resolution**:
- Verify `CLICKHOUSE_ADDR` and credentials in Grafana's datasource provisioning files.
- Ensure ClickHouse pods are running and passing readiness probes.

## 5. MinIO Disk Full
**Symptoms**: ClickHouse fails to move data to the warm tier. ClickHouse local disk fills up.
**Diagnosis**:
- Check MinIO console or metrics for disk usage.
**Resolution**:
- Expand the Persistent Volume Claim (PVC) for MinIO.
- Adjust the ClickHouse TTL policy to drop data earlier if storage is strictly constrained.
- Implement MinIO lifecycle policies to move older data to a cheaper tier or delete it.

## 6. Migration Failures
**Symptoms**: Go API fails to start, complaining about database schema.
**Diagnosis**:
- Read API logs during startup.
**Resolution**:
- Manual intervention required. Exec into the ClickHouse client, check the `schema_migrations` table, and manually apply or fix the failing SQL script.
