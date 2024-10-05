#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
commit=$(tr -d '[:space:]' < "$root/grammars/pinned-commit.txt")
tab=$(printf '\t')
while IFS="$tab" read -r lang exts pkg dir main extra base transform patch; do
  case "$lang" in
    lang|"") continue ;;
  esac
  dst="$root/grammars/$lang"
  mkdir -p "$dst"
  base_url="https://raw.githubusercontent.com/antlr/grammars-v4/$commit/$dir"
  curl -fsSL -o "$dst/$main" "$base_url/$main"
  if [ "$extra" != "-" ] && [ -n "$extra" ]; then
    curl -fsSL -o "$dst/$extra" "$base_url/$extra"
  fi
  if [ "$base" != "-" ] && [ -n "$base" ]; then
    oldifs=$IFS
    IFS=,
    for f in $base; do
      curl -fsSL -o "$dst/$f.txt" "$base_url/Go/$f"
    done
    IFS=$oldifs
  fi
done < "$root/grammars/languages.tsv"
