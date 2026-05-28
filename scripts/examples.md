## Usage

```
get-repos.sh <project> [--output-dir <path>] [--force-cleanup]
```

- `<project>` - project to clone/fetch (e.g. `rhtas`, `rhtas-releases`)
- `--output-dir <path>` - override the output directory from config
- `--force-cleanup` - discard local changes before checkout

## Branch resolution order

1. Per-repo `branch` field in config
2. Project-level `default_branch` in config

If neither is set, the tool exits with an error.

## Configuration

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

## Examples

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
