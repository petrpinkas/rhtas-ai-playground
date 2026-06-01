#!/usr/bin/env bash
# Finds open konflux/mintmaker docker-deps PRs targeting a release branch
# across all git repositories in a given directory.

set -euo pipefail

usage() {
  echo "Usage: $0 <release-branch> [directory]" >&2
  echo "  release-branch  required — base branch and mintmaker head prefix (e.g. release-1.3)" >&2
  echo "  directory       directory containing git repos (default: .)" >&2
  exit 1
}

[ $# -lt 1 ] && usage

BRANCH="$1"
DIR="${2:-.}"
MINTMAKER_HEAD="konflux/mintmaker/${BRANCH}/docker-deps"

if [ ! -d "$DIR" ]; then
  echo "Error: '$DIR' is not a directory" >&2
  exit 1
fi

for cmd in gh jq; do
  if ! command -v "$cmd" &>/dev/null; then
    echo "Error: '$cmd' is required but not found." >&2
    exit 1
  fi
done

found=0
missing=()

for repo_dir in "$DIR"/*/; do
  [[ -d "$repo_dir/.git" ]] || continue

  remote=$(git -C "$repo_dir" remote get-url origin 2>/dev/null) || continue
  # gh pr list only applies to GitHub; skip GitLab and other hosts (e.g. konflux-release-data).
  [[ "$remote" == *github.com* ]] || continue

  repo_slug=$(echo "$remote" | sed -E 's|\.git$||; s|.*[:/]([^/]+/[^/]+)$|\1|')

  result=$(gh pr list \
    --repo "$repo_slug" \
    --base "$BRANCH" \
    --state open \
    --head "${MINTMAKER_HEAD}" \
    --json number,title,url 2>/dev/null) || result=""

  if [[ -n "$result" && "$result" != "[]" ]]; then
    found=$((found + 1))
    echo "$(basename "$repo_dir")"
    echo "$result" | jq -r '.[] | "  #\(.number) \(.title)\n  \(.url)"'
  else
    missing+=("$(basename "$repo_dir")")
  fi
done

echo ""
echo "Found docker-deps PRs in ${found} repo(s)."

if [[ ${#missing[@]} -gt 0 ]]; then
  echo "No PR in: ${missing[*]}"
fi
