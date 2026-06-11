#!/usr/bin/env bash
# Fail if merged Go coverage is below MIN_COVERAGE (default 94) on hand-written code.
# Generated protobuf (packages/proto/gen/go) is excluded — see TEST_ARCHITECTURE.md.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MIN_RAW="${MIN_COVERAGE:-94}"
if ! [[ "$MIN_RAW" =~ ^[0-9]+$ ]]; then
  echo "MIN_COVERAGE must be an integer, got: $MIN_RAW"
  exit 1
fi
MIN="$MIN_RAW"
PROFILE="${1:-coverage-go-merged.out}"

if [[ ! -f "$PROFILE" ]]; then
  echo "coverage profile not found: $PROFILE"
  exit 1
fi

FILTERED="${PROFILE%.out}-handwritten.out"
bash "$ROOT/infra/scripts/coverage-filter.sh" "$PROFILE" "$FILTERED"

PCT=$(go tool cover -func="$FILTERED" | awk '/^total:/ { gsub(/%/,"",$3); print $3 }')
echo "hand-written coverage: ${PCT}% (minimum ${MIN}%, excludes packages/proto/gen/go and infra/)"

awk -v pct="$PCT" -v min="$MIN" 'BEGIN { exit (pct + 0 >= min + 0) ? 0 : 1 }'
