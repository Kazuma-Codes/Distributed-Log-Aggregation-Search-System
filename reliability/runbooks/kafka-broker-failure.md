# Kafka Broker Failure

## Overview
A Kafka broker in the `log-kafka` cluster has failed, become unresponsive, or crashed. This runbook covers single broker failure, multi-broker failure, split brain, and disk failures.

## Severity
- **Single Broker**: Warning (Cluster should continue operating normally due to replication)
- **Multi-Broker**: Critical (Potential data loss or ingestion pipeline halt)

## Symptoms
- Alerts fire: `KafkaBrokerDown`, `KafkaUnderReplicatedPartitions`
- Users experience: Potential brief latency spikes during leader election.
- Metrics show: Drop in active controller count, missing broker in `kafka_server_replicamanager_activebrokercount`.

## Diagnosis
1. Check pod status:
   `kubectl get pods -n log-platform -l app.kubernetes.io/name=kafka`
2. Check broker logs for the specific failed pod:
   `kubectl logs -n log-platform log-kafka-kafka-0 -c kafka --tail=100`
3. Describe pod to check for OOM or Eviction:
   `kubectl describe pod log-kafka-kafka-0 -n log-platform`
4. Verify Under Replicated Partitions (URP):
   `kubectl exec -it log-kafka-kafka-1 -n log-platform -c kafka -- bin/kafka-topics.sh --describe --under-replicated-partitions --bootstrap-server localhost:9092`

## Remediation

### Single Broker Failure
1. The Strimzi operator should automatically restart the pod. Wait up to 5 minutes.
2. If the pod is in CrashLoopBackOff due to a disk space issue, proceed to the disk failure section.
3. If it's a transient node issue, delete the pod to force a restart on another node:
   `kubectl delete pod log-kafka-kafka-0 -n log-platform`

### Disk Failure
1. If PVC is full, see `storage-full.md`.
2. If disk is corrupted, delete the PVC associated with the broker and restart the pod. Strimzi will recreate it and replicate data:
   `kubectl delete pvc data-log-kafka-kafka-0 -n log-platform`
   `kubectl delete pod log-kafka-kafka-0 -n log-platform`

### Multi-Broker Failure / Split Brain
1. Halt producers temporarily if possible.
2. Check ZooKeeper / KRaft quorum status.
3. Restart brokers sequentially. Verify ISR catches up before moving to the next.

## Prevention
- Ensure multi-AZ deployment with Pod Anti-Affinity.
- Adjust memory limits and requests based on profiling.
- Ensure disk alarms are set at 80% to give sufficient reaction time.

## Escalation
Escalate to SRE Lead if more than one broker is down or if URP persists for over 1 hour.
