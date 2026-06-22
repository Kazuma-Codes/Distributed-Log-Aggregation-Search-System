# Deployment Guide

This guide covers deploying the Distributed Log Aggregation & Search System across different environments.

## Prerequisites
- **Docker & Docker Compose** (for local development)
- **Kubernetes Cluster** (v1.24+)
- **kubectl** & **Helm** (v3+)
- **Go 1.23** (for building the API)

## Local Development (Docker Compose)
The easiest way to run the stack locally is using Docker Compose.

1. **Copy the environment file**:
   ```bash
   cp .env.example .env
   ```
2. **Start the services**:
   ```bash
   make compose-up
   ```
   This starts Kafka, ClickHouse, Vector, MinIO, Grafana, and Prometheus.
3. **Verify the deployment**:
   ```bash
   docker-compose ps
   ```
4. **Stop the services**:
   ```bash
   make compose-down
   ```

## Kubernetes Deployment
We use a GitOps-friendly approach, leveraging Operators where applicable.

### 1. Setup K3d Cluster (Local K8s)
```bash
make cluster-create
```

### 2. Deploy Operators
We rely on the Strimzi Operator for Kafka and the ClickHouse Operator.
```bash
helm repo add strimzi https://strimzi.io/charts/
helm install strimzi-kafka-operator strimzi/strimzi-kafka-operator --namespace kafka --create-namespace
```

### 3. Deploy Components
Deploy components in the following order to ensure dependencies are met:

```bash
# 1. MinIO (Storage Backend)
make deploy-minio

# 2. Kafka (Message Broker)
make deploy-kafka

# 3. ClickHouse (Database)
make deploy-clickhouse

# 4. Vector (Log Collector)
make deploy-vector

# 5. API & Observability
make deploy-api
make deploy-grafana
make deploy-prometheus
```
*Alternatively, run `make deploy-all`.*

## Cloud Deployment (AWS/GCP)
When deploying to managed cloud providers:
- **Storage**: Swap MinIO for native AWS S3 or Google Cloud Storage (GCS).
- **Compute**: Use managed node groups. Assign node taints for ClickHouse nodes to ensure high-I/O instances (e.g., AWS `i3` or `im4gn` instances) are used exclusively for the database.
- **Kafka**: Consider MSK (AWS) or Confluent Cloud if you prefer not to manage Strimzi yourself.

## Configuration Reference
All components read from environment variables or ConfigMaps. Refer to `.env.example` for the core application variables.
- **ClickHouse**: Configured via `clickhouse-config.yaml` injected by the Operator.
- **Vector**: Configured via `vector.toml` ConfigMap.

## Post-Deployment Verification Checklist
- [ ] Kafka brokers are `Ready` and the `logs` topic exists.
- [ ] ClickHouse cluster is `Ready` and `Kafka` table engine is consuming.
- [ ] Vector pods are running on all nodes (`DaemonSet`).
- [ ] Go API pods are healthy and responding to `/health`.
- [ ] Grafana datasources (ClickHouse, Prometheus) are connected.

## Rollback Procedures
If a deployment fails:
1. **API/Vector**: Use standard Kubernetes rollbacks: `kubectl rollout undo deployment/log-api`
2. **Kafka/ClickHouse**: Changes managed by Operators should be reverted in the Custom Resource definitions. Do NOT manually delete pods.
