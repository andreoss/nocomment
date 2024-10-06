#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
mkdir -p scratch
profile=scratch/cover.out
go test -coverprofile="$profile" ./internal/cli ./internal/filter ./internal/lang ./internal/langs ./internal/lexer ./internal/process ./internal/scan ./internal/conformance >/dev/null
line=$(go tool cover -func="$profile" | tail -1)
percent=$(printf '%s\n' "$line" | grep -oE '[0-9]+\.[0-9]+' | tail -1)
echo "coverage ${percent}%"
rm -f "$profile"
awk -v p="$percent" 'BEGIN { if (p + 0 < 85) exit 1 }'
