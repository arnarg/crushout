#!/bin/bash
set -euo pipefail

crushout="${1:?usage: run.sh <crushout-bin> <cases.jsonl>}"
cases="${2:?usage: run.sh <crushout-bin> <cases.jsonl>}"

passed=0
failed=0
lineno=0

while IFS= read -r line; do
  lineno=$((lineno + 1))

  name=$(echo "$line" | jq -r '.name')
  input=$(echo "$line" | jq '.input')
  expected=$(echo "$line" | jq -c '.expected')

  config=$(echo "$line" | jq -r '.config // empty')
  if [ -n "$config" ]; then
    config_dir=$(mktemp -d)
    printf '%b' "$config" > "$config_dir/.crushout.yml"
    input=$(echo "$input" | jq --arg cwd "$config_dir" '.cwd = $cwd')
  fi

  actual=$(echo "$input" | "$crushout" 2>/dev/null) || true
  actual_compact=$(echo "$actual" | jq -c '.')

  if [ "$actual_compact" = "$expected" ]; then
    passed=$((passed + 1))
  else
    failed=$((failed + 1))
    echo "FAIL (line $lineno): $name"
    echo "  expected: $expected"
    echo "  actual:   $actual_compact"
  fi

  if [ -n "$config" ] && [ -d "$config_dir" ]; then
    rm -rf "$config_dir"
  fi
done < "$cases"

echo "$passed passed, $failed failed"
[ "$failed" -eq 0 ]
