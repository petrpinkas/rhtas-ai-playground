# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this project does

A tool to batch-clone/fetch and checkout multiple Git repositories defined in a YAML config. Used to manage sets of RHTAS (Red Hat Trusted Artifact Signer) repositories from the `securesign` GitHub org across different release branches.

## Build and run

```bash
# Build (from scripts-go/)
go build ./cmd/get-repos/

# Run via wrapper (from project root)
./scripts/get-repos.sh <project> [--output-dir <path>] [--force-cleanup]

# Run directly (from scripts-go/)
go run ./cmd/get-repos rhtas
```

No tests exist. Single dependency: `gopkg.in/yaml.v3`.

## Architecture

Single-file Go CLI at `scripts-go/cmd/get-repos/main.go`. Config at `scripts-go/config/repositories.yaml`.

**Config structure:**
- `repositories` — prefix aliases mapping short names to git URL bases (e.g. `securesign` -> `git@github.com:securesign`)
- `projects` — named sets of repos, each with `dir` (output directory), `default_branch`, and a list of `repos` (each with `url` and optional `branch` override)

**Key functions:**
- `resolveURL` — expands `securesign/cosign` to `git@github.com:securesign/cosign.git` using the `repositories` map
- `resolveBranch` — per-repo `branch` field, then project `default_branch`; errors if neither is set
- `repoName` — extracts directory name from URL (last path segment without `.git`)

**Flow:** parse args -> load config -> resolve URLs and branches -> for each repo: clone or fetch -> checkout branch -> report local changes and failures.

## Config notes

- `dir` is required per project (or pass `--output-dir`). Supports relative and absolute paths.
- Branch must always be configured — there is no fallback to `main`.
- Repo URLs use prefix aliases from `repositories` section; full git URLs also work as-is.
