#!/bin/bash
set -euo pipefail

echo "Starting port-forwarding for all UIs..."

kubectl port-forward svc/grafana 3000:3000 -n log-platform &
P1=$!

kubectl port-forward svc/chi-log-clickhouse-logs-0-0 8123:8123 -n log-platform &
P2=$!

kubectl port-forward svc/minio 9001:9001 -n log-platform &
P3=$!

kubectl port-forward svc/prometheus 9090:9090 -n log-platform &
P4=$!

# Assuming a query-api service is deployed
kubectl port-forward svc/query-api 8080:8080 -n log-platform 2>/dev/null &
P5=$!

function cleanup() {
    echo "Killing all port-forwards..."
    kill $P1 $P2 $P3 $P4 $P5 2>/dev/null || true
    wait
    exit 0
}

trap cleanup SIGINT SIGTERM

echo "Port forwarding active. Press Ctrl+C to stop."
wait
