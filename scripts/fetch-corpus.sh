#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
manifest="$root/internal/conformance/testdata/manifest.tsv"
dest="$root/internal/conformance/testdata/files"
mkdir -p "$dest"
tab=$(printf '\t')
count=0
skipped=0
while IFS="$tab" read -r language repo commit path license; do
  case "$language" in
    language|"") continue ;;
  esac
  out="$dest/$language"
  mkdir -p "$out"
  name=$(basename "$path")
  if curl -fsSL -o "$out/$name" "https://raw.githubusercontent.com/$repo/$commit/$path"; then
    count=$((count + 1))
  else
    rm -f "$out/$name"
    skipped=$((skipped + 1))
    echo "skip $language $path" >&2
  fi
done < "$manifest"
echo "fetched $count, skipped $skipped"
