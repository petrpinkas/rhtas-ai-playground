#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: $(basename "$0") <repo-slug> [--days N]" >&2
  echo "  repo-slug  GitHub repo (e.g. securesign/cosign)" >&2
  echo "  --days N   look back N days (default: 7)" >&2
  exit 1
}

SLUG=""
DAYS=7

while [[ $# -gt 0 ]]; do
  case "$1" in
    --days)
      DAYS="$2"; shift 2 ;;
    --help|-h)
      usage ;;
    *)
      SLUG="$1"; shift ;;
  esac
done

[[ -z "$SLUG" ]] && usage

command -v gh &>/dev/null || { echo "Error: 'gh' CLI is required" >&2; exit 1; }
command -v jq &>/dev/null || { echo "Error: 'jq' is required" >&2; exit 1; }

since=$(date -v-"${DAYS}"d +%Y-%m-%d 2>/dev/null || date -d "${DAYS} days ago" +%Y-%m-%d)

prs=$(gh pr list \
  --repo "$SLUG" \
  --state merged \
  --search "merged:>=${since}" \
  --json number,title,mergedAt,mergedBy,author,url,baseRefName,headRefName \
  --limit 100)

count=$(echo "$prs" | jq 'length')

echo "=== $SLUG — ${count} PR(s) merged since $since ($DAYS days) ==="

if [[ "$count" -eq 0 ]]; then
  exit 0
fi

pr_numbers=$(echo "$prs" | jq -r 'sort_by(.mergedAt) | reverse | .[].number')

for pr_number in $pr_numbers; do
  pr_json=$(echo "$prs" | jq ".[] | select(.number == ${pr_number})")

  title=$(echo "$pr_json" | jq -r '.title')
  pr_author=$(echo "$pr_json" | jq -r '.author.login')
  merged_by=$(echo "$pr_json" | jq -r '.mergedBy.login')
  merged_at=$(echo "$pr_json" | jq -r '.mergedAt')
  head_ref=$(echo "$pr_json" | jq -r '.headRefName')
  base_ref=$(echo "$pr_json" | jq -r '.baseRefName')
  url=$(echo "$pr_json" | jq -r '.url')

  echo ""
  echo "PR #${pr_number}: ${title}"
  echo "  Author:    ${pr_author}"
  echo "  Merged by: ${merged_by}"
  echo "  Merged at: ${merged_at}"
  echo "  Branch:    ${head_ref} -> ${base_ref}"
  echo "  URL:       ${url}"

  commits=$(gh api "repos/${SLUG}/pulls/${pr_number}/commits" --paginate 2>/dev/null || echo "[]")
  commit_count=$(echo "$commits" | jq 'length')
  commit_authors=$(echo "$commits" | jq -r '[.[].commit.author.name] | unique | join(", ")')
  echo "  Commits:   ${commit_count} by ${commit_authors}"

  files=$(gh api "repos/${SLUG}/pulls/${pr_number}/files" --paginate 2>/dev/null || echo "[]")
  echo "  Files:"
  echo "$files" | jq -r '.[] | "    \(.status | gsub("added";"A") | gsub("removed";"D") | gsub("modified";"M") | gsub("renamed";"R") | gsub("copied";"C") | gsub("changed";"M"))  \(.filename)"'
done
