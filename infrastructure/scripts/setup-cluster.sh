#!/bin/bash
set -euo pipefail

CLUSTER_NAME="log-platform"

echo "Creating k3d cluster: $CLUSTER_NAME..."
if k3d cluster list | grep -q "$CLUSTER_NAME"; then
    echo "Cluster $CLUSTER_NAME already exists."
else
    k3d cluster create "$CLUSTER_NAME" \
      --servers 1 \
      --agents 3 \
      -p "9092:9092@loadbalancer" \
      -p "9094:9094@loadbalancer" \
      -p "3000:3000@loadbalancer" \
      -p "8123:8123@loadbalancer" \
      -p "9000:9000@loadbalancer" \
      -p "9001:9001@loadbalancer" \
      -p "9090:9090@loadbalancer" \
      --wait
fi

echo "Creating log-platform namespace..."
kubectl apply -f ../k8s/namespace.yaml

echo "Cluster setup complete."
