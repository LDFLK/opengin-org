#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=== Step 1: Link minister organisations ==="
go run "$SCRIPT_DIR/link_minister_orgs/main.go"

# echo ""
# echo "=== Waiting 3 minutes before step 2 ==="
# sleep 120

echo ""
echo "=== Step 2: Link citizen roles ==="
go run "$SCRIPT_DIR/link_citizen_roles/main.go"

echo ""
echo "=== Done ==="
