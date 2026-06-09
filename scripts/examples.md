## get-repos

```
get-repos.sh <project> [--output-dir <path>] [--force-cleanup]
```

- `<project>` - project to clone/fetch (e.g. `rhtas`, `rhtas-releases`)
- `--output-dir <path>` - override the output directory from config
- `--force-cleanup` - discard local changes before checkout

### Branch resolution order

1. Per-repo `branch` field in config
2. Project-level `default_branch` in config

If neither is set, the tool exits with an error.

### Configuration

Defined in `scripts-go/config/repositories.yaml`.

The `repositories` section maps short prefixes to git URL bases, so repos can be referenced concisely in projects:

```yaml
repositories:
  securesign: git@github.com:securesign

projects:
  rhtas:
    dir: ../repositories/rhtas          # relative path
    default_branch: main
    repos:
      - url: securesign/cosign          # expands to git@github.com:securesign/cosign.git
      - url: securesign/tough
        branch: develop                  # per-repo branch override
  rhtas-abs:
    dir: /home/user/projects/rhtas      # absolute path
    default_branch: main
    repos:
      - url: securesign/cosign
```

Full git URLs also work directly (bypassing the prefix lookup):

```yaml
repos:
  - url: git@github.com:other-org/some-repo.git
```

### Examples

```bash
# Clone/fetch all repos in the rhtas project using configured dir and branches
./scripts/get-repos.sh rhtas

# Clone/fetch all repos in rhtas-releases
./scripts/get-repos.sh rhtas-releases

# Override the output directory (relative or absolute)
./scripts/get-repos.sh rhtas --output-dir ./my-repos
./scripts/get-repos.sh rhtas --output-dir /tmp/rhtas-repos

# Discard local changes before checkout
./scripts/get-repos.sh rhtas --force-cleanup
```

## list-merged-prs

Lists merged PRs for a single GitHub repository, including author, merger, commits, and changed files.

```
list-merged-prs.sh <repo-slug> [--days N]
```

- `<repo-slug>` - GitHub repository (e.g. `securesign/cosign`)
- `--days N` - look back N days (default: 7)

### Examples

```bash
# List PRs merged in securesign/cosign in the last 7 days
./scripts/list-merged-prs.sh securesign/cosign

# Last 30 days
./scripts/list-merged-prs.sh securesign/cosign --days 30
```

## list-merged-prs-all

Runs `list-merged-prs.sh` on all securesign repositories. The repo list is defined at the top of the script.

```
list-merged-prs-all.sh [--days N]
```

- `--days N` - look back N days (default: 7)

### Examples

```bash
# All securesign repos, last 7 days
./scripts/list-merged-prs-all.sh

# All securesign repos, last 3 days
./scripts/list-merged-prs-all.sh --days 3
```

## check-local-changes

Scans a directory of git repositories for uncommitted changes.

```
check-local-changes.sh [directory]
```

- `directory` - directory containing git repos (default: `.`)

### Examples

```bash
./scripts/check-local-changes.sh repos
```

## check-mintmaker-docker-prs

Finds open Konflux/mintmaker docker-deps PRs targeting a release branch across all git repositories in a directory.

```
check-mintmaker-docker-prs.sh <release-branch> [directory]
```

- `<release-branch>` - base branch and mintmaker head prefix (e.g. `release-1.3`)
- `directory` - directory containing git repos (default: `.`)

### Examples

```bash
./scripts/check-mintmaker-docker-prs.sh release-1.3 repos
```
