#!/bin/bash

# Step 1: Creates an organisation structure for each minister with the following structure:
# Minister A -> Organisation -> Minister -> Secretary

# Step 2: links the existing people in the db to the Minister node in the organisation structure
# using AS_ROLE relationship.

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=== Step 1: Create the org structure ==="
go run "$SCRIPT_DIR/create_org_structure/main.go"


echo ""
echo "=== Step 2: Link minister roles ==="
go run "$SCRIPT_DIR/link_minister_roles/main.go"

echo ""
echo "=== Done ==="
