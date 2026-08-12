#!/usr/bin/env bash
#
# Creates rhtas-v1.0.2 tags for policy-controller and policy-controller-operator.
# Commits derived from Konflux snapshots in snapshot.json (2026-08-02 build).
# FBC excluded per request.
#
# Usage:
#   ./create-tags-pco-1.0.2.sh           # dry run (default, safe)
#   ./create-tags-pco-1.0.2.sh --apply   # actually create the tags
#

set -euo pipefail

ORG="securesign"
DRY_RUN=true

if [[ "${1:-}" == "--apply" ]]; then
  DRY_RUN=false
fi

# Format: "repo|tag|full_commit_hash|source_snapshot"
TAGS=(
  "policy-controller|rhtas-v1.0.2|2c76b571169d66ec3c4d3cee0d47b375fc61cd8c|policy-controller-v1-0-20260730-125814-000"
  "policy-controller-operator|rhtas-v1.0.2|f60ef2ad6dd3dccc523198f971c46c089e6a0f30|policy-controller-operator-v1-0-20260805-151337-000"
)

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

if [[ "$DRY_RUN" == "true" ]]; then
  echo -e "${YELLOW}DRY RUN — no tags will be created. Pass --apply to create them.${NC}"
  echo
fi

SKIPPED=0
CREATED=0
FAILED=0

for entry in "${TAGS[@]}"; do
  IFS='|' read -r repo tag commit snapshot <<< "$entry"

  printf "%-35s %-18s %s ... " "$repo" "$tag" "${commit:0:9}"

  if gh api "repos/$ORG/$repo/git/ref/tags/$tag" > /dev/null 2>&1; then
    echo -e "${YELLOW}SKIP (already exists)${NC}"
    (( SKIPPED++ )) || true
    continue
  fi

  if [[ "$DRY_RUN" == "true" ]]; then
    echo -e "${GREEN}would create${NC}  (from $snapshot)"
    (( CREATED++ )) || true
    continue
  fi

  if gh api "repos/$ORG/$repo/git/refs" \
      --method POST \
      --field "ref=refs/tags/$tag" \
      --field "sha=$commit" \
      > /dev/null 2>&1; then
    echo -e "${GREEN}OK${NC}  (from $snapshot)"
    (( CREATED++ )) || true
  else
    echo -e "${RED}FAILED${NC}"
    (( FAILED++ )) || true
  fi
done

echo
if [[ "$DRY_RUN" == "true" ]]; then
  echo "Would create: $CREATED  |  Already exist: $SKIPPED"
else
  echo -e "Created: ${GREEN}$CREATED${NC}  |  Skipped: ${YELLOW}$SKIPPED${NC}  |  Failed: ${RED}$FAILED${NC}"
fi
