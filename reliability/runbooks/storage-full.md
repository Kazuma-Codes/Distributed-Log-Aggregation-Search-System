# Storage Full

## Overview
One of the stateful components (ClickHouse, Kafka, or MinIO) has exhausted its allocated disk space, risking service failure or data loss.

## Severity
Critical

## Symptoms
- Alerts fire: `DiskSpaceUsageHigh`, `KafkaDiskFull`, `ClickHouseDiskFull`, `MinioDiskFull`
- Users experience: Ingestion failures (500s or timeouts), inability to query data.
- Metrics show: Volume usage > 90% or 95%.

## Diagnosis
1. **Identify the affected component:**
   `kubectl get pvc -n log-platform`
   Use standard monitoring (Prometheus/Grafana) to check disk utilization metrics.
2. **Kafka:**
   Check partition sizes:
   `kubectl exec -n log-platform log-kafka-kafka-0 -c kafka -- du -sh /var/lib/kafka/data/kafka-log*`
3. **ClickHouse:**
   Check disk space by disk and table:
   `kubectl exec -n log-platform chi-log-clickhouse-logs-0-0-0 -- clickhouse-client -q "SELECT name, path, free_space, total_space FROM system.disks"`
4. **MinIO:**
   Check MinIO usage through the MinIO UI or `mc` CLI.

## Remediation

### Kafka Disk Full
1. **Reduce Retention:** Temporarily reduce the retention period or size limit for large topics.
   `kubectl exec -n log-platform log-kafka-kafka-0 -c kafka -- bin/kafka-configs.sh --bootstrap-server localhost:9092 --alter --entity-type topics --entity-name logs --add-config retention.ms=86400000`
2. **Increase PVC Size:** If the StorageClass supports volume expansion, edit the PVC to request more storage.

### ClickHouse Disk Full (Hot Storage)
1. **Force TTL Moves:** Ensure TTL rules are working to move data from local SSDs to MinIO (cold storage).
   `OPTIMIZE TABLE logs_local FINAL`
2. **Emergency Cleanup:** Drop old partitions manually if TTL is broken.
   `ALTER TABLE logs_local DROP PARTITION 'YYYYMMDD'`
3. **Expand PVC:** Expand local storage if supported.

### MinIO Disk Full (Cold Storage)
1. **Adjust ILM/TTL:** Adjust object lifecycle policies in MinIO to expire older data faster.
2. **Add Nodes/Drives:** Expand the MinIO cluster by adding more nodes or increasing disk sizes.
3. **Delete Old Data:** Manually delete older log buckets or prefixes.

## Prevention
- Ensure aggressive TTLs are in place for logs.
- Set up automated volume expansion if supported by the cloud provider.
- Set alerts for 70% and 80% disk usage to provide ample warning.

## Escalation
Escalate to Infrastructure/SRE Lead if disks reach 95% and automated/quick remediation fails, requiring manual data destruction.
