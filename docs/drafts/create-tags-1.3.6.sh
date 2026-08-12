#!/usr/bin/env bash
#
# Creates rhtas-v1.3.6 tags in all securesign repositories.
# Commits derived from Konflux snapshots in snapshot.json (2026-07-14 build).
#
# Usage:
#   ./create-tags-1.3.6.sh           # dry run (default, safe)
#   ./create-tags-1.3.6.sh --apply   # actually create the tags
#

set -euo pipefail

ORG="securesign"
DRY_RUN=true

if [[ "${1:-}" == "--apply" ]]; then
  DRY_RUN=false
fi

# Format: "repo|tag|full_commit_hash|source_snapshot"
# Conflicts resolved by newest commit date.
# fbc excluded pending v4-22 commit clarification — add manually if needed.
TAGS=(
  "secure-sign-operator|v1.3.6|f79dcca477144e3548e960d14738349242be7170|operator-v1-3-20260714-162137-000"
  "secure-sign-operator|rhtas-v1.3.6|f79dcca477144e3548e960d14738349242be7170|operator-v1-3-20260714-162137-000"
  "certificate-transparency-go|rhtas-v1.3.6|ca6ad62159d0fee273b7740912de33ceae5331b6|tas-components-v1-3-20260714-131258-000-n"
  "fulcio|rhtas-v1.3.6|c54e7c67d07a8f49ae8ce43be0048392b54b2eff|tas-components-v1-3-20260714-131258-000-n"
  "rekor-search-ui|rhtas-v1.3.6|ad001a115ba918aa25397d64ac152d4861df7b19|tas-components-v1-3-20260714-131258-000-n"
  "rekor|rhtas-v1.3.6|7c13df8730eb6fd593e73afdb3a92ac774f55c6a|tas-components-v1-3-20260714-131258-000-n"
  "timestamp-authority|rhtas-v1.3.6|b0511150434ad8559f5ea217947dd08b22dd90b8|tas-components-v1-3-20260714-131258-000-n"
  "trillian|rhtas-v1.3.6|04a40b115eae343e6d9e939dd732ef3ea31ff36c|tas-components-v1-3-20260714-131258-000-n (newest of 3 commits)"
  "cosign|rhtas-v1.3.6|8caa10231d6c333ebb14834190c7cc11790fa056|client-server-v1-3-20260714-142713-000 (newest)"
  "gitsign|rhtas-v1.3.6|dbf936ef41fe1f920aac5d80ccd2f0924b3f5bc1|tas-tools-v1-3-20260714-132503-000-n"
  "tough|rhtas-v1.3.6|a74e46c4028cfb85d375bfd8e7dc570954e792d2|tough-v1-3-20260714-083649-000"
  "rekor-monitor|rhtas-v1.3.6|4dafce7e1b7d4a183e6cbf958b127ba30f069fb0|rekor-monitor-v1-3-20260713-184903-000"
  "segment-backup-job|rhtas-v1.3.6|16b86a4105f037330fd0bbe43481301423ea05d7|segment-backup-job-v1-3-20260714-081156-000"
  "artifact-signer-ansible|rhtas-v1.3.6|d1a418e7c4282e83a17d1a51c20bafeab6144a80|ansible-v1-3-20260714-173941-000"
  "fbc|rhtas-v1.3.6|1f40dab46dc7860b003102e7e352acf9d4082d72|rhtas-fbc-20260714-185804-000 (v4-16..v4-21)"
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

  # Check if tag already exists
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
