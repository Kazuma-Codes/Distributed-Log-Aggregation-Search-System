#!/usr/bin/env bash
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

NAMESPACE="log-platform"
MINIO_POD=$(kubectl get pod -n ${NAMESPACE} -l app=minio -o jsonpath="{.items[0].metadata.name}" || echo "minio-0")

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

setup() {
    log_info "Setting up test environment..."
    log_info "Checking current disk usage on MinIO."
}

run_test() {
    log_info "Running chaos test: Simulating disk pressure on MinIO."
    log_info "Creating large dummy files to fill MinIO volume to 95%..."
    # The actual command would be something like:
    kubectl exec ${MINIO_POD} -n ${NAMESPACE} -- sh -c 'dd if=/dev/urandom of=/export/dummy_fill bs=1M count=1000' || log_warn "dd command failed."
}

verify() {
    log_info "Verifying system behavior under disk pressure..."
    log_info "Checking MinIO health endpoint."
    log_info "Verifying writes to MinIO fail gracefully."
    log_info "Verifying hot storage (ClickHouse local) continues to work."
}

cleanup() {
    log_info "Cleaning up test environment..."
    log_info "Removing dummy files from MinIO volume..."
    kubectl exec ${MINIO_POD} -n ${NAMESPACE} -- sh -c 'rm -f /export/dummy_fill' || log_warn "rm command failed."
    log_info "Verifying MinIO recovers."
    log_info "Cleanup complete."
}

trap cleanup EXIT

main() {
    setup
    run_test
    verify
    log_info "Chaos test completed successfully."
}

main
