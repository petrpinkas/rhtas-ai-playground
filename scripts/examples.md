- [check-local-changes](#check-local-changes)
- [check-mintmaker-docker-prs](#check-mintmaker-docker-prs)
- [check-tags-n-branches](#check-tags-n-branches)
- [get-projects](#get-projects)
- [list-merged-prs](#list-merged-prs)
- [list-merged-prs-all](#list-merged-prs-all)

## check-local-changes

Scans a directory of git repositories for uncommitted changes.

```
check-local-changes [directory]
```

- `directory` - directory containing git repos (default: `.`)

### Examples

```bash
./scripts/check-local-changes repos
```

## check-mintmaker-docker-prs

Finds open Konflux/mintmaker docker-deps PRs targeting a release branch across all git repositories in a directory.

```
check-mintmaker-docker-prs <release-branch> [directory]
```

- `<release-branch>` - base branch and mintmaker head prefix (e.g. `release-1.3`)
- `directory` - directory containing git repos (default: `.`)

### Examples

```bash
./scripts/check-mintmaker-docker-prs release-1.3 repos
```

## check-tags-n-branches

Checks whether expected release branches and tags exist in remote repositories, based on release groups defined in the config.

```
check-tags-n-branches [--repo <name>]... [--branch <name>]... [--tag <name>]...
                      [--branches-only] [--tags-only]
```

- `--repo <name>` - check only this repo (short name or full prefix/name)
- `--branch <name>` - check only this branch
- `--tag <name>` - check only this tag
- `--branches-only` - skip tag checks
- `--tags-only` - skip branch checks

All filters can be repeated and match against what is defined in the config. With no flags, checks everything across all groups.

### Configuration

Uses the `release_groups` section in `scripts-go/config/repositories.yaml`. Each group defines a set of repos and the branches/tags expected to exist in them:

```yaml
release_groups:
  securesign:
    release_branches:
      - release-1.3
      - release-1.4
    release_tags:
      - rhtas-v1.3.5
      - rhtas-v1.4.1
    repos:
      - securesign/cosign
      - securesign/fulcio

  tags-only:
    release_tags:
      - rhtas-v1.3.5
    repos:
      - securesign/fbc
```

Repo URLs use prefix aliases from the `repositories` section (same as `get-projects`). Groups without `release_branches` or `release_tags` are skipped.

### Examples

```bash
# Check all repos, branches, and tags across all release groups
./scripts/check-tags-n-branches

# Check a single repo
./scripts/check-tags-n-branches --repo cosign

# Check only a specific branch across all repos
./scripts/check-tags-n-branches --branch release-1.3

# Check only branches (skip tags)
./scripts/check-tags-n-branches --branches-only

```

## get-projects

```
get-projects <project> [--output-dir <path>] [--force-cleanup]
```

- `<project>` - project to clone/fetch (e.g. `rhtas`, `rhtas-releases`)
- `--output-dir <path>` - output directory (default: current directory)
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
    default_branch: main
    repos:
      - url: securesign/cosign          # expands to git@github.com:securesign/cosign.git
      - url: securesign/tough
        branch: develop                  # per-repo branch override
```

Full git URLs also work directly (bypassing the prefix lookup):

```yaml
repos:
  - url: git@github.com:other-org/some-repo.git
```

### Examples

```bash
# Clone/fetch all repos in the rhtas project into the current directory
./scripts/get-projects rhtas

# Clone/fetch all repos in rhtas-releases
./scripts/get-projects rhtas-releases

# Override the output directory (relative or absolute)
./scripts/get-projects rhtas --output-dir ./my-repos
./scripts/get-projects rhtas --output-dir /tmp/rhtas-repos

# Discard local changes before checkout
./scripts/get-projects rhtas --force-cleanup
```

## list-merged-prs

Lists merged PRs for a single GitHub repository, including author, merger, commits, and changed files.

```
list-merged-prs <repo-slug> [--days N]
```

- `<repo-slug>` - GitHub repository (e.g. `securesign/cosign`)
- `--days N` - look back N days (default: 7)

### Examples

```bash
# List PRs merged in securesign/cosign in the last 7 days
./scripts/list-merged-prs securesign/cosign

# Last 30 days
./scripts/list-merged-prs securesign/cosign --days 30
```

## list-merged-prs-all

Runs `list-merged-prs` on all securesign repositories. The repo list is defined at the top of the script.

```
list-merged-prs-all [--days N]
```

- `--days N` - look back N days (default: 7)

### Examples

```bash
# All securesign repos, last 7 days
./scripts/list-merged-prs-all

# All securesign repos, last 3 days
./scripts/list-merged-prs-all --days 3
```
