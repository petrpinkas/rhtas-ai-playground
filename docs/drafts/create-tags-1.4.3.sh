#!/usr/bin/env bash
#
# Creates rhtas-v1.4.3 / v1.4.3 tags in all securesign repositories.
# Commits derived from Konflux snapshots in snapshot.json (2026-07-30 build).
#
# Usage:
#   ./create-tags-1.4.3.sh           # dry run (default, safe)
#   ./create-tags-1.4.3.sh --apply   # actually create the tags
#

set -euo pipefail

ORG="securesign"
DRY_RUN=true

if [[ "${1:-}" == "--apply" ]]; then
  DRY_RUN=false
fi

# Format: "repo|tag|full_commit_hash|source_snapshot"
TAGS=(
  "secure-sign-operator|v1.4.3|d397f009aa14476eb9d06ff77d258a70bc1225e9|operator-v1-4-20260730-110157-000-9m"
  "secure-sign-operator|rhtas-v1.4.3|d397f009aa14476eb9d06ff77d258a70bc1225e9|operator-v1-4-20260730-110157-000-9m"
  "certificate-transparency-go|rhtas-v1.4.3|3fab3e55e7838024fa88a1656bca870365d2d4fc|tas-components-v1-4-20260730-062129-000"
  "fulcio|rhtas-v1.4.3|03ee1680663a1327f958050cecee70f45697e33c|tas-components-v1-4-20260730-062129-000"
  "rekor-search-ui|rhtas-v1.4.3|e5739ea586ac948c0c33565ca1a49e636fbd872c|tas-components-v1-4-20260730-062129-000"
  "rekor-monitor|rhtas-v1.4.3|5189ed3930149940febaa977fc1cc1dd8b507661|rekor-monitor-v1-4-20260730-061538-000"
  "artifact-signer-ansible|rhtas-v1.4.3|65ed2f30ac330471e9a293cdcc27e35154951bdb|ansible-v1-4-20260730-190426-000"
  "cosign|rhtas-v1.4.3|f48ed4f0a57ee59c7813f450cc7daa52eb1f1406|client-server-v1-4-20260730-074512-000"
  "rekor|rhtas-v1.4.3|55171081d055fd15bf531ce18ac28136e67baef8|cli-stacks-v1-4-20260730-072727-000-b"
  "gitsign|rhtas-v1.4.3|cacbed9f75c4b0ac8b11891c30bc04ec26d3898e|cli-stacks-v1-4-20260730-072727-000-b"
  "timestamp-authority|rhtas-v1.4.3|f42b17485b11b96eac7d6b9001cda30d2070bd0f|cli-stacks-v1-4-20260730-072727-000-b"
  "tough|rhtas-v1.4.3|9195b03ac20bf477e8836af7c975fda40486a5d5|cli-stacks-v1-4-20260730-072727-000-b"
  "trillian|rhtas-v1.4.3|5e3ffa09ee112c0f37aff41678291ee2ef2640ce|cli-stacks-v1-4-20260730-072727-000-b"
  "fbc|rhtas-v1.4.3|41284d4173bfaeeceb510803a5551f62d32dc1d2|rhtas-fbc-20260730-111908-000-1a (v4-17..v4-22)"
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
