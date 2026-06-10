#!/usr/bin/env bash
# Strip generated protobuf packages from a Go coverprofile (hand-written scope only).
set -euo pipefail

INPUT="${1:?usage: coverage-filter.sh INPUT.out [OUTPUT.out]}"
OUTPUT="${2:-${INPUT%.out}-handwritten.out}"

awk '
  /^mode:/ { print; next }
  /packages\/proto\/gen\/go\// { next }
  { print }
' "$INPUT" > "$OUTPUT"

echo "filtered profile: $OUTPUT ($(wc -l < "$OUTPUT" | tr -d ' ') lines)"
