# ADR 002: ClickHouse over Elasticsearch

**Status:** Accepted

## Context
We require a storage and query engine capable of handling billions of log entries. The system must support real-time ingestion of 100K+ logs/sec, retain data cost-effectively, and provide sub-second latency for aggregations and filtering. The industry standard has traditionally been Elasticsearch (the ELK stack), but we evaluated ClickHouse as an alternative.

## Decision
We decided to use **ClickHouse** as the primary datastore and query engine.

## Reasons
1. **Performance**: ClickHouse is an OLAP column-oriented database. For aggregations (e.g., counting errors over time) and filtering, it is often 10-100x faster than Elasticsearch.
2. **Storage Efficiency**: ClickHouse offers extremely high compression ratios (up to 10x better than Elasticsearch) due to its columnar format and advanced codecs (e.g., ZSTD). This dramatically reduces storage costs.
3. **Licensing**: ClickHouse is licensed under Apache 2.0. Elasticsearch changed its license to the Server Side Public License (SSPL), which is not OSI-approved.
4. **Native Integration**: ClickHouse has a native Kafka Table Engine, allowing it to pull messages directly from Kafka without needing an intermediary service like Logstash or an external consumer app.
5. **SQL Interface**: ClickHouse uses SQL, making it accessible to a wider range of developers compared to Elasticsearch's custom JSON query DSL.

## Consequences
- **Full-Text Search**: ClickHouse is not a specialized search engine. Its full-text search capabilities are less mature than Elasticsearch. However, by leveraging Bloom filters (`tokenbf_v1`), skip indexes, and careful partitioning, it is highly efficient for typical log searching patterns.
- **Ecosystem**: The community tooling for logs specific to ClickHouse is smaller than the massive ELK ecosystem, requiring some custom dashboarding in Grafana and custom API development.
