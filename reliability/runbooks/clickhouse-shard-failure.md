# ClickHouse Shard Failure

## Overview
A ClickHouse shard in the `log-clickhouse` cluster is unreachable, crashing, or failing to accept queries. This covers single shard failure, replica sync issues, stuck merges, and OOM kills.

## Severity
- **Replica Failure**: Warning (Queries and inserts fall back to healthy replicas)
- **All Replicas in Shard Down**: Critical (Partial data loss for queries, insert errors)

## Symptoms
- Alerts fire: `ClickHouseNodeDown`, `ClickHouseReplicationLag`
- Users experience: Slower queries, incomplete results, or HTTP 500s from the query API.
- Metrics show: `ClickHouseProfileEvents_InsertedRows` dropping, increased query latencies.

## Diagnosis
1. Check pod status:
   `kubectl get pods -n log-platform -l app=clickhouse`
2. Check pod logs:
   `kubectl logs -n log-platform chi-log-clickhouse-logs-0-0-0 --tail=100`
3. Connect to a healthy instance and check cluster status:
   `kubectl exec -it chi-log-clickhouse-logs-0-1-0 -n log-platform -- clickhouse-client -q "SELECT cluster, shard_num, replica_num, host_name, is_local FROM system.clusters"`
4. Check for replica sync issues (mutations/merges):
   `kubectl exec -it chi-log-clickhouse-logs-0-1-0 -n log-platform -- clickhouse-client -q "SELECT * FROM system.replication_queue"`

## Remediation

### Single Shard / Replica Failure
1. Pod Restart: If a pod is OOMKilled, it will be restarted by Kubernetes. Check memory limits if it happens repeatedly.
2. If the pod is stuck, forcefully delete it:
   `kubectl delete pod chi-log-clickhouse-logs-0-0-0 -n log-platform`

### Replica Sync Issues / Stuck Merges
1. If the replication queue is stuck, you may need to drop the problematic partition on the affected replica and let it sync from the good replica:
   `ALTER TABLE logs_local DROP PARTITION 'YYYYMMDD'`
2. Verify ZooKeeper health, as ClickHouse relies on it heavily for replication. Restart ZooKeeper pods if they are out of quorum.

### OOM Kills
1. An expensive query might be consuming all RAM. Identify the query:
   `SELECT query FROM system.query_log ORDER BY memory_usage DESC LIMIT 5`
2. Kill the offending query:
   `KILL QUERY WHERE query_id = '...'`
3. Increase `max_memory_usage` or add resource limits to specific users.

## Prevention
- Optimize ClickHouse queries to avoid large memory footprints.
- Implement quota and query complexity limits.
- Scale out ClickHouse by adding more shards if resources are consistently exhausted.

## Escalation
Escalate to SRE Lead and Data Engineering if cluster quorum is lost or if queries are failing globally.
