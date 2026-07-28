# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this project does

A tool to batch-clone/fetch and checkout multiple Git repositories defined in a YAML config. Used to manage sets of RHTAS (Red Hat Trusted Artifact Signer) repositories from the `securesign` GitHub org across different release branches.

## Build and run

```bash
# Build (from scripts-go/)
go build ./cmd/get-projects/

# Initialize a project in the current directory (creates .projects.yaml)
./scripts/get-projects init rhtas

# Update repos from .projects.yaml in the current directory
./scripts/get-projects [--force-cleanup]
```

No tests exist. Single dependency: `gopkg.in/yaml.v3`.

## Architecture

Single-file Go CLI at `scripts-go/cmd/get-projects/main.go`. Config at `scripts-go/config/repositories.yaml`.

**Two-step workflow:**
1. `get-projects init <project>` — copies the named project's config from `repositories.yaml` into `.projects.yaml` in the current directory
2. `get-projects` — reads `.projects.yaml` from the current directory, clones/fetches repos there

**Config structure (repositories.yaml):**
- `repositories` — prefix aliases mapping short names to git URL bases (e.g. `securesign` -> `git@github.com:securesign`)
- `projects` — named sets of repos, each with `default_branch` and a list of `repos` (each with `url` and optional `branch` override)

**Local config (.projects.yaml):** flat structure with `repositories` (only used prefixes), `default_branch`, and `repos`.

**Key functions:**
- `runInit` — reads bundled config, extracts project + used prefixes, writes `.projects.yaml`
- `runUpdate` — reads `.projects.yaml`, runs clone/fetch/checkout cycle
- `resolveURL` — expands `securesign/cosign` to `git@github.com:securesign/cosign.git` using the `repositories` map
- `repoName` — extracts directory name from URL (last path segment without `.git`)

## Config notes

- Repos are cloned into the current directory (where `.projects.yaml` lives).
- Branch must always be configured — there is no fallback to `main`.
- Repo URLs use prefix aliases from `repositories` section; full git URLs also work as-is.
- `init` refuses to overwrite an existing `.projects.yaml`; delete it first to re-initialize.
