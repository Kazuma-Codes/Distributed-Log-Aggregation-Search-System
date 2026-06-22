# Benchmarking Methodology

This document outlines how we benchmark the Log Platform to ensure it meets its performance targets (100K+ logs/sec, sub-second query latency).

## Test Environment Specifications
All formal benchmarks should be run on a standardized cloud environment to ensure reproducibility.

**Target Environment (AWS):**
- **ClickHouse**: 3x `i3.4xlarge` (16 vCPU, 122GB RAM, 2x 1.9TB NVMe)
- **Kafka**: 3x `m5.2xlarge` (8 vCPU, 32GB RAM, io1 EBS volumes)
- **Vector Aggregators**: 2x `c5.2xlarge`
- **Load Generators**: 5x `c5.xlarge`

## Load Generation Approach
We use **k6** combined with a custom Go script to generate log traffic.
- **Generator**: A custom Go binary (`make generate-logs`) generates realistic JSON logs (Nginx access logs, App JSON logs with stack traces).
- **Injection**: Logs are sent via HTTP to the Vector Aggregator or directly to the API endpoint to simulate different ingress paths.
- **Scale**: We ramp up virtual users in k6 to incrementally increase logs/sec until the system breaks or SLA is breached.

## Metrics Collection
During the test, we collect metrics via Prometheus:
1. **Ingestion Rate**: `sum(rate(vector_events_out_total[1m]))`
2. **Kafka Latency**: 99th percentile of produce request latency.
3. **Consumer Lag**: `kafka_consumergroup_lag`
4. **ClickHouse Insert Rate**: Rows per second inserted into the `MergeTree`.
5. **Query Latency**: API HTTP response times (p50, p95, p99).
6. **Resource Usage**: CPU, Memory, Disk I/O across all nodes.

## Statistical Methodology
- **Warm-up**: The system is run at 10K logs/sec for 10 minutes to populate caches and stabilize JVMs (Kafka).
- **Sustained Load**: The target load (e.g., 100K logs/sec) is maintained for 1 hour.
- **Query Mix**: During sustained load, a script executes a mix of queries:
  - 70% simple point lookups (e.g., `trace_id = 'X'`).
  - 20% aggregations (e.g., `count by level over 1h`).
  - 10% full-text scans (e.g., `message ILIKE '%exception%'`).

## Reproducibility Instructions
To run a basic local benchmark:
1. Start the cluster: `make compose-up`
2. Run the load test: `make load-test` (Ensure k6 is installed).
3. Monitor the Grafana `Benchmark` dashboard.
4. Export the results: `k6 run scripts/load_test.js --out json=results.json`.
