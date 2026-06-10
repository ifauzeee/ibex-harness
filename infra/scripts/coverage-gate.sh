#!/usr/bin/env bash
# Fail if merged Go coverage is below MIN_COVERAGE (default 94) on hand-written code.
# Generated protobuf (packages/proto/gen/go) is excluded — see TEST_ARCHITECTURE.md.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MIN="${MIN_COVERAGE:-94}"
PROFILE="${1:-coverage-go-merged.out}"

if [[ ! -f "$PROFILE" ]]; then
  echo "coverage profile not found: $PROFILE"
  exit 1
fi

FILTERED="${PROFILE%.out}-handwritten.out"
bash "$ROOT/infra/scripts/coverage-filter.sh" "$PROFILE" "$FILTERED"

PCT=$(go tool cover -func="$FILTERED" | awk '/^total:/ { gsub(/%/,"",$3); print $3 }')
echo "hand-written coverage: ${PCT}% (minimum ${MIN}%, excludes packages/proto/gen/go)"

awk -v pct="$PCT" -v min="$MIN" 'BEGIN { exit (pct + 0 >= min + 0) ? 0 : 1 }'
