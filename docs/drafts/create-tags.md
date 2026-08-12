# Creating Release Tags

Process for creating git tags in securesign repositories for a given RHTAS (or related product) release version.

## Prerequisites

- `oc` logged in to the Konflux cluster (`oc whoami` should return your username)
- `gh` authenticated to GitHub (`gh auth status`)
- The release `snapshot.json` file for the target version

## Step 1 — Find the snapshot.json

Snapshot files live under the releases repo checkout:

```
releases-main/<product>/<operator>/<version>/stable/snapshot.json
```

Examples:
- RHTAS 1.4.3: `releases-main/rhtas/rhtas-operator/1.4.3/stable/snapshot.json`
- PCO 1.0.2: `releases-main/policy-controller/policy-controller-operator/1.0.2/stable/snapshot.json`

The file lists all Konflux snapshot names used in the release, grouped by application (e.g. `tas-components-v1-4`, `operator-v1-4`, `rhtas-fbc`, …).

## Step 2 — Query Konflux for source repos and commits

For each snapshot name in the file, query Konflux to get the authoritative source URL and revision for every component:

```bash
oc get snapshot <snapshot-name> -n rhtas-tenant -o json | \
  jq -r '.spec.components[] | "\(.name) | \(.source.git.url) | \(.source.git.revision)"'
```

Run this for all snapshot names in parallel. This is more reliable than reading `vcs-ref` labels from the container images because it is the data Konflux itself recorded at build time.

**Key observations:**
- Components that come from the same source repo will share the same revision.
- A single snapshot (e.g. `tas-components`) can contain components from many different repos.
- Some repos appear in multiple snapshots (e.g. `cosign` appears in both `tas-tools` and `client-server`) — they will have different commits.

## Step 3 — Resolve conflicts (multiple commits per repo)

When a repo appears in more than one snapshot at different commits, pick the **newest** commit by checking the commit date via the GitHub API:

```bash
gh api repos/securesign/<repo>/commits/<sha> --jq '.commit.committer.date'
```

Run all lookups in parallel. The newest commit is the correct one to tag.

**Known cases where conflicts occur:**
- `cosign` — appears in `tas-tools` (cosign-cli) and `client-server`
- `trillian` — `database` component often lags behind the other trillian components
- `rekor` — appears in `tas-components` and `cli-stacks`
- `gitsign`, `timestamp-authority`, `tough` — appear in both main snapshots and `cli-stacks`

## Step 4 — Determine tag format

All repos use the `rhtas-v<version>` format. `secure-sign-operator` additionally gets a plain `v<version>` tag, but the `rhtas-v<version>` tag is the one that matters and must always be created.

## Step 5 — Handle FBC separately

The `fbc` (and `pco-fbc`) snapshot often contains components at different commits for different OCP versions. Check whether most OCP versions share the same commit:

```bash
oc get snapshot <fbc-snapshot-name> -n rhtas-tenant -o json | \
  jq -r '.spec.components[] | "\(.name) | \(.source.git.revision)"'
```

Use the commit shared by the majority of OCP versions. If one version has an outlier commit, verify with the team whether that OCP version was actually released before tagging. Exclude FBC from the main script and add it explicitly once confirmed.

## Step 6 — Create the tagging script

Generate a shell script following this structure:

```bash
#!/usr/bin/env bash
set -euo pipefail

ORG="securesign"
TAG="rhtas-v1.4.3"
DRY_RUN=true
[[ "${1:-}" == "--apply" ]] && DRY_RUN=false

# Format: "repo|full_commit_hash|source_snapshot"
TAGS=(
  "secure-sign-operator|<sha>|operator-v1-4-..."
  "fulcio|<sha>|tas-components-v1-4-..."
  # ... one entry per repo
)

for entry in "${TAGS[@]}"; do
  IFS='|' read -r repo commit snapshot <<< "$entry"
  printf "%-35s %-18s %s ... " "$repo" "$TAG" "${commit:0:9}"

  if gh api "repos/$ORG/$repo/git/ref/tags/$TAG" > /dev/null 2>&1; then
    echo "SKIP (already exists)"
    continue
  fi

  if [[ "$DRY_RUN" == "true" ]]; then
    echo "would create  (from $snapshot)"
    continue
  fi

  gh api "repos/$ORG/$repo/git/refs" \
    --method POST \
    --field "ref=refs/tags/$TAG" \
    --field "sha=$commit" > /dev/null \
    && echo "OK" || echo "FAILED"
done
```

**Important:** the tag existence check must use the exit code of `gh api`, not its output — a 404 response still produces output that looks non-empty when piped through `jq`.

## Step 7 — Dry run, then apply

Always run without `--apply` first to confirm what will be created vs skipped:

```bash
./create-tags-<version>.sh          # dry run
./create-tags-<version>.sh --apply  # create tags
```

Tags are created via the GitHub API — no local clones needed.

## Special cases

- **Repo not in the snapshot**: If a repo appears in `tags-and-branches` as `[MISSING]` but has no corresponding image in the snapshot, find the tag manually (e.g. by locating the existing tag of the same format from a prior release and checking if a newer one exists on the release branch).
- **Already-tagged repos with a second format**: When a `v<version>` tag already exists and you need to add `rhtas-v<version>`, retrieve the commit from the existing tag: `gh api repos/securesign/<repo>/git/ref/tags/v<version> --jq '.object.sha'`, then create the new tag pointing to the same commit.
