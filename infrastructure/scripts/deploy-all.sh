#!/bin/bash
set -euo pipefail

cd "$(dirname "$0")/.."

echo "1. Creating namespace..."
kubectl apply -f k8s/namespace.yaml

echo "2. Installing Strimzi Operator..."
kubectl apply -k k8s/kafka/
echo "Waiting for Strimzi Operator to be ready..."
sleep 15
kubectl wait deployment/strimzi-cluster-operator --for=condition=available --timeout=300s -n log-platform || true

echo "Deploying Kafka Cluster..."
kubectl apply -f k8s/kafka/kafka-cluster.yaml
echo "Waiting for Kafka to be ready..."
kubectl wait kafka/log-kafka --for=condition=Ready --timeout=300s -n log-platform || echo "Kafka still provisioning"

echo "Deploying Kafka Topics..."
kubectl apply -f k8s/kafka/kafka-topics.yaml

echo "3. Installing Altinity ClickHouse Operator..."
kubectl apply -k k8s/clickhouse/
echo "Waiting for ClickHouse Operator to be ready..."
sleep 15
kubectl wait deployment/clickhouse-operator --for=condition=available --timeout=300s -n log-platform || true

echo "Deploying ClickHouse Cluster..."
kubectl apply -f k8s/clickhouse/storage-policy.yaml
kubectl apply -f k8s/clickhouse/clickhouse-cluster.yaml

echo "4. Deploying MinIO..."
kubectl apply -f k8s/minio/minio-deployment.yaml
kubectl apply -f k8s/minio/minio-service.yaml
echo "Waiting for MinIO to be ready..."
kubectl wait pod -l app.kubernetes.io/name=minio --for=condition=Ready --timeout=300s -n log-platform || true

echo "Initializing MinIO buckets..."
kubectl apply -f k8s/minio/bucket-init.yaml

echo "5. Deploying Vector..."
kubectl apply -f k8s/vector/vector-rbac.yaml
kubectl apply -f k8s/vector/vector-configmap.yaml
kubectl apply -f k8s/vector/vector-daemonset.yaml

echo "6. Deploying Prometheus & AlertManager..."
kubectl apply -f k8s/prometheus/prometheus-configmap.yaml
kubectl apply -f k8s/prometheus/prometheus-deployment.yaml
kubectl apply -f k8s/prometheus/alertmanager.yaml

echo "7. Deploying Grafana..."
kubectl apply -f k8s/grafana/grafana-configmap.yaml
kubectl apply -f k8s/grafana/grafana-dashboards-cm.yaml
kubectl apply -f k8s/grafana/grafana-deployment.yaml

echo "Deployment complete."
