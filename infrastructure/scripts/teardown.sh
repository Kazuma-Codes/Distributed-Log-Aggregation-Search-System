#!/bin/bash
set -euo pipefail

read -p "Are you sure you want to tear down the log-platform namespace? This will destroy all data! (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Deleting namespace log-platform..."
    kubectl delete namespace log-platform --ignore-not-found=true
    echo "Teardown complete."
else
    echo "Teardown aborted."
fi
