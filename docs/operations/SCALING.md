# Scaling Playbook

This document provides guidelines for scaling the Log Platform to handle increased throughput or query loads.

## Horizontal Scaling

### 1. Vector (Log Collector)
- **Mechanism**: Deployed as a DaemonSet. It scales automatically as you add nodes to your Kubernetes cluster.
- **Aggregators**: If using Vector Aggregators, use HPA (Horizontal Pod Autoscaler) based on CPU utilization (target 70%).

### 2. Kafka (Buffer)
- **Mechanism**: Add more brokers to the cluster via the Strimzi `Kafka` Custom Resource.
- **Partitions**: You MUST increase the number of partitions for the `logs` topic to distribute the load across new brokers and allow more parallel ClickHouse consumers.
- **Formula**: `Target Partitions = Target Throughput / Max Throughput per Partition (typically 10-20MB/s)`.

### 3. ClickHouse (Storage & Queries)
- **Write Scaling**: Add more **Shards**. This distributes the data and write load across more nodes.
- **Read Scaling**: Add more **Replicas** to existing shards. This allows parallel execution of read queries.
- **Mechanism**: Update the `ClickHouseInstallation` CRD. The Operator handles node provisioning, but data rebalancing across new shards must be triggered manually.

### 4. Go API
- **Mechanism**: Standard Kubernetes HPA.
- **Triggers**: Scale based on CPU (target 60%) or custom Prometheus metrics (e.g., HTTP request rate).

## Vertical Scaling

- **ClickHouse**: Responds very well to vertical scaling. Giving ClickHouse more RAM allows for larger in-memory aggregations and faster group-bys. More CPU cores linearly improve query speed for complex searches.
- **Kafka**: Generally bound by Network and Disk I/O. Use instance types with high network bandwidth and provision fast NVMe SSDs or high-IOPS EBS volumes.

## Bottleneck Identification

| Component | Metric to Watch | Threshold indicating Bottleneck |
| :--- | :--- | :--- |
| **Vector** | `vector_buffer_byte_size` | Nearing max configured buffer size |
| **Kafka** | `kafka_server_brokertopicmetrics_bytesinpersec` | Reaching network interface limit |
| **ClickHouse Consumer** | Kafka Consumer Lag | Consistently > 0 and growing |
| **ClickHouse DB** | `ClickHouseProfileEvents_Query` | High latency / High CPU utilization |

## Capacity Planning Formulas

**Storage Estimation:**
```
Daily Raw Volume = (Average Log Size in Bytes) * (Logs per Second) * 86400
Daily Compressed Volume = Daily Raw Volume / Compression Ratio (assume 8x to 10x)
Hot Storage Needs = Daily Compressed Volume * 7 days * Replication Factor (e.g., 2)
```

**Cost Estimation considerations:**
- Node compute (EC2/GKE instances).
- EBS/NVMe storage (Hot tier).
- S3 storage (Warm/Cold tier).
- Cross-AZ network traffic (Replication).
