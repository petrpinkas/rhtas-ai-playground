#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../scripts-go"
go run ./cmd/check-tags-n-branches "$@"
