# High Ingestion Lag

## Overview
Data is arriving at the system but not appearing in ClickHouse within the expected latency bounds. This can stem from Kafka consumer lag, Vector backpressure, or a bottleneck in ClickHouse inserts.

## Severity
Critical

## Symptoms
- Alerts fire: `KafkaConsumerLagHigh`, `VectorBufferHigh`
- Users experience: Missing recent logs in the dashboard.
- Metrics show: Growing consumer lag, vector buffer utilization > 80%, high ClickHouse insert latencies.

## Diagnosis
1. **Check Kafka Consumer Lag:**
   `kubectl exec -n log-platform log-kafka-kafka-0 -c kafka -- bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group vector-consumer`
2. **Check Vector Health:**
   `kubectl get pods -n log-platform -l app=vector`
   Check logs for `Vector` experiencing backpressure or connection timeouts to ClickHouse.
3. **Check ClickHouse Insert Bottleneck:**
   Check ClickHouse metrics for `system.metrics` where metric is `InsertQueries` or check for 'Too many parts' errors.
   `kubectl exec -n log-platform chi-log-clickhouse-logs-0-0-0 -- clickhouse-client -q "SELECT * FROM system.errors WHERE last_error_time > now() - INTERVAL 1 HOUR"`

## Remediation

### Upstream Spike (Legitimate Traffic Increase)
1. Scale up Vector pods:
   `kubectl scale deployment vector -n log-platform --replicas=10`
2. Increase Kafka partitions if consumers are maxed out (Vector concurrency).

### ClickHouse Insert Bottleneck ("Too many parts")
1. Vector is sending batches that are too small or too frequently. Adjust Vector configuration to increase `batch.max_bytes` and `batch.timeout_secs`.
2. Check if ClickHouse merges are falling behind. If so, reduce the insert rate temporarily or scale ClickHouse vertically.

### Vector Backpressure
1. If Vector is queueing due to an unreachable sink, verify ClickHouse network connectivity.
2. If disk buffer is full on Vector, you may need to clear the buffer if the data is stale, or allocate larger disks to Vector PVCs.

## Prevention
- Ensure Kafka topics are over-partitioned to allow easy horizontal scaling of consumers.
- Tune Vector batching to optimize for ClickHouse's preference for large, infrequent inserts (e.g., 100k+ rows per insert).
- Implement rate limiting at the API gateway for log ingestion.

## Escalation
Escalate to SRE Lead if lag continues to grow after scaling components or if there is a risk of Kafka log retention periods expiring before consumption.
