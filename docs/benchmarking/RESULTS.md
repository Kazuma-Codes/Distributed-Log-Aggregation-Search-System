# Benchmark Results: v1.0.0

*Date: October 2023*
*Environment: AWS (See METHODOLOGY.md for spec)*

## Summary Table

| Metric | Target | Result | Status |
| :--- | :--- | :--- | :--- |
| **Max Ingestion Throughput** | 100,000 logs/sec | 125,000 logs/sec | ✅ PASS |
| **P99 Query Latency (1h)** | < 500 ms | 120 ms | ✅ PASS |
| **P99 Query Latency (24h)** | < 1,000 ms | 450 ms | ✅ PASS |
| **Storage Compression** | 10x | 12.5x | ✅ PASS |
| **Kafka Buffer Size** | 24h retention | Held 3TB data | ✅ PASS |

## Detailed Results

### 1. Ingestion Performance
- **Peak Sustained Load**: Reached 125K logs/sec (approx 150 MB/sec).
- **Bottleneck Hit**: At 130K logs/sec, Kafka disk I/O on `m5.2xlarge` became the bottleneck, causing producer latency to spike and Vector buffers to fill.
- **ClickHouse Performance**: The ClickHouse Kafka Engine easily kept up with 125K logs/sec using 4 consumer threads per node. CPU utilization on ClickHouse remained under 40%.

### 2. Query Performance
Queries were executed concurrently with a background ingestion rate of 50K logs/sec.

* **Point Lookup (Indexed ID):**
  * P50: 15 ms
  * P99: 45 ms
* **Aggregation (Count errors grouped by service over 24h):**
  * P50: 120 ms
  * P99: 450 ms
* **Full-Text Scan (Regex matching over 1h without skip index):**
  * P50: 800 ms
  * P99: 1,200 ms *(Note: Missed target, requires index optimization)*

### 3. Storage Efficiency
- **Raw JSON Volume**: 1.5 TB generated over the test duration.
- **ClickHouse On-Disk Volume**: 120 GB.
- **Compression Ratio**: 12.5x. This was achieved using the `ZSTD(1)` codec on the payload columns and dictionary encoding for low-cardinality fields (e.g., `level`, `service_name`).

## Graphs

*(Insert grafana screenshot of ingestion rate holding steady at 125K)*
**Figure 1: Ingestion Rate over 1 Hour**

*(Insert grafana screenshot of API request latency)*
**Figure 2: API Query Latency Distribution**

## Comparison with Targets
We successfully met or exceeded all primary targets. The architecture proves capable of handling enterprise-scale logging workloads. The only area requiring attention is full-text regex scanning on non-indexed columns.

## Recommendations
1. **Kafka Upgrades**: To push beyond 125K logs/sec, upgrade Kafka brokers to instances with higher baseline EBS bandwidth or use locally attached NVMe (e.g., `i3en` instances).
2. **ClickHouse Indexes**: Implement `tokenbf_v1` bloom filters on the raw `message` column to improve the latency of generic text searches.
3. **Warm Tier Testing**: Future benchmarks must evaluate query performance when data is pulled from the MinIO S3 warm tier.
