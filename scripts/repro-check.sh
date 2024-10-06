#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
hash_tree() {
  find internal/lexer/generated -type f | sort | while read -r file; do
    sha256sum "$file"
  done | sha256sum | awk '{print $1}'
}
./scripts/generate-lexers.sh >/dev/null 2>&1
first=$(hash_tree)
./scripts/generate-lexers.sh >/dev/null 2>&1
second=$(hash_tree)
if [ "$first" != "$second" ]; then
  echo "generated tree not reproducible" >&2
  exit 1
fi
echo "generated tree reproducible: $first"
