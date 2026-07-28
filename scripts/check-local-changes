#!/usr/bin/env bash
set -euo pipefail

dir="${1:-.}"

if [ ! -d "$dir" ]; then
  echo "Error: $dir is not a directory" >&2
  exit 1
fi

any_changes=false

while IFS= read -r git_dir; do
  repo_dir="$(dirname "$git_dir")"
  branch="$(git -C "$repo_dir" branch --show-current 2>/dev/null)"
  [ -z "$branch" ] && branch="detached"
  status="$(git -C "$repo_dir" status --short 2>/dev/null)" || continue
  if [ -n "$status" ]; then
    any_changes=true
    echo "--- $repo_dir (on $branch) --- CHANGES ---"
    echo "$status"
    echo
  else
    echo "--- $repo_dir (on $branch) ---"
  fi
done < <(find "$dir" -type d -name .git | sort)

if [ "$any_changes" = false ]; then
  echo
  echo "No local changes found in any repository under $dir"
fi
