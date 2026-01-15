#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GIT_DIR="$(cd "$REPO_ROOT" && git rev-parse --git-dir)"

mkdir -p "$GIT_DIR/hooks"
ln -sf "$SCRIPT_DIR/pre-commit" "$GIT_DIR/hooks/pre-commit"
echo "Pre-commit hook installed"
