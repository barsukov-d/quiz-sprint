#!/bin/bash

# PostToolUse hook: run vue-tsc type-check after editing .ts/.tsx/.vue files in tma/
# Input: JSON from stdin with tool_input.file_path

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty' 2>/dev/null)

# Only run for TypeScript/Vue files inside tma/
if [[ "$file_path" == *"/tma/"* ]] && [[ "$file_path" =~ \.(ts|tsx|vue)$ ]]; then
  cd "$CLAUDE_PROJECT_DIR/tma" || exit 0
  pnpm run type-check 2>&1
fi

exit 0
