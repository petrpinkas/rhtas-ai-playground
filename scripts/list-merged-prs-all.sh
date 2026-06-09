#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(dirname "$0")"

REPOS=(
  securesign/Snapshot-Reporter
  securesign/actions
  securesign/artifact-signer-ansible
  securesign/certificate-transparency-go
  securesign/community-operators-prod
  securesign/cosign
  securesign/demo-resources
  securesign/e2e-gitsign-test
  securesign/fbc
  securesign/fulcio
  securesign/gitsign
  securesign/go-version-package
  securesign/helm-charts
  securesign/helper-scripts
  securesign/hive-config
  securesign/image-factory
  securesign/integration-tests
  securesign/model-transparency
  securesign/model-transparency-go
  securesign/model-validation-operator
  securesign/pipelines
  securesign/pipelines-demo
  securesign/policy-controller
  securesign/policy-controller-operator
  securesign/quickstarts
  securesign/rekor
  securesign/rekor-monitor
  securesign/rekor-search-ui
  securesign/releases
  securesign/renovate-config
  securesign/rhtas-benchmark
  securesign/rhtas-console
  securesign/rhtas-console-ui
  securesign/rhtas-dashboard
  securesign/scaffolding
  securesign/secure-sign-operator
  securesign/segment-backup-job
  securesign/sigstore-a2a
  securesign/sigstore-a2a-go
  securesign/sigstore-ansible
  securesign/sigstore-e2e
  securesign/sigstore-ocp
  securesign/sigstore-python
  securesign/sigstore-rs
  securesign/structural-tests
  securesign/team-docs
  securesign/timestamp-authority
  securesign/tough
  securesign/trillian
  securesign/trusted-artifact-signer-hcc-ui
  securesign/trusted-foundations
  securesign/tufcli
)

usage() {
  echo "Usage: $(basename "$0") [--days N]" >&2
  echo "  --days N   look back N days (default: 7)" >&2
  exit 1
}

DAYS=7

while [[ $# -gt 0 ]]; do
  case "$1" in
    --days)
      DAYS="$2"; shift 2 ;;
    --help|-h)
      usage ;;
    *)
      echo "Error: unknown argument '$1'" >&2; usage ;;
  esac
done

for repo in "${REPOS[@]}"; do
  "$SCRIPT_DIR/list-merged-prs.sh" "$repo" --days "$DAYS"
  echo ""
done
