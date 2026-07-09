#!/bin/bash
set -euo pipefail

NAMESPACE="rhtas-tenant"
ANNOTATION="test.appstudio.openshift.io/keep-snapshot=true"
ANNOTATION_KEY="test.appstudio.openshift.io/keep-snapshot"

if [[ $# -ne 1 ]]; then
    echo "Usage: $0 <path-to-snapshot.json>"
    exit 1
fi

SNAPSHOT_FILE="$1"

if [[ ! -f "$SNAPSHOT_FILE" ]]; then
    echo "Error: file not found: $SNAPSHOT_FILE"
    exit 1
fi

snapshot_names=$(jq -r '.. | .snapshot_name? // empty' "$SNAPSHOT_FILE")

already_annotated=$(oc get snapshots.appstudio.redhat.com -n "$NAMESPACE" -o json \
    | jq -r '.items[]
        | select(.metadata.annotations["'"$ANNOTATION_KEY"'"] == "true")
        | .metadata.name')

for name in $snapshot_names; do
    if echo "$already_annotated" | grep -qx "$name"; then
        echo "SKIP  $name (already annotated)"
    else
        echo "ANNOTATE  $name"
        oc annotate snapshot "$name" -n "$NAMESPACE" "$ANNOTATION"
    fi
done
