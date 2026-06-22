# ADR 004: Three-Tier Storage Strategy

**Status:** Accepted

## Context
Log data volume grows rapidly, and keeping all historical logs on high-performance storage (like NVMe SSDs) becomes prohibitively expensive. However, different ages of logs have vastly different access patterns:
- Last 24 hours: 90% of queries (troubleshooting active incidents).
- 1 to 7 days: 9% of queries (recent trends, weekly reports).
- > 7 days: 1% of queries (compliance, historical audits).

## Decision
We decided to implement a **Three-Tier Storage Strategy** using ClickHouse's tiered storage capabilities backed by MinIO.

- **Hot Tier**: ClickHouse local NVMe/SSDs for logs up to 7 days old.
- **Warm Tier**: MinIO (S3 Standard) for logs between 7 and 90 days old.
- **Cold Tier**: MinIO (S3 Glacier/Deep Archive equivalent) for logs older than 90 days.

## Reasons
1. **Cost Efficiency**: Object storage (MinIO/S3) is orders of magnitude cheaper than block storage (EBS/NVMe). Moving older data allows us to retain massive volumes of logs without breaking the budget.
2. **Transparent Querying**: ClickHouse can automatically move data between its disks using TTL policies. Queries against older data are automatically routed to read from the object storage, requiring no changes to the application or queries.
3. **Diminishing Returns**: Fast storage for 30-day old data provides little value since it's rarely queried.

## Consequences
- **Query Latency**: Queries that span beyond the 7-day window will be significantly slower, as data must be fetched over the network from MinIO instead of read from local disk.
- **Complexity**: Requires managing a MinIO cluster and configuring ClickHouse storage policies and TTL expressions.
- **Data Lifecycle Management**: We must implement strict bucket lifecycle rules in MinIO to move data from warm to cold, and eventually expire it.
