#!/bin/sh
# Flush pending bd (beads) issue changes before commit

if ! command -v bd >/dev/null 2>&1; then
  exit 0
fi

BEADS_DIR=""
if git rev-parse --git-dir >/dev/null 2>&1; then
  if [ "$(git rev-parse --git-dir)" != "$(git rev-parse --git-common-dir)" ]; then
    MAIN_REPO_ROOT="$(dirname "$(git rev-parse --git-common-dir)")"
    [ -d "$MAIN_REPO_ROOT/.beads" ] && BEADS_DIR="$MAIN_REPO_ROOT/.beads"
  else
    [ -d .beads ] && BEADS_DIR=".beads"
  fi
fi

[ -z "$BEADS_DIR" ] && exit 0

if ! bd sync --flush-only >/dev/null 2>&1; then
  echo "Error: Failed to flush bd changes. Run 'bd sync --flush-only' to diagnose." >&2
  exit 1
fi

if [ -f "$BEADS_DIR/issues.jsonl" ]; then
  if [ "$(git rev-parse --git-dir)" = "$(git rev-parse --git-common-dir)" ]; then
    git add "$BEADS_DIR/issues.jsonl" 2>/dev/null || true
  fi
fi
