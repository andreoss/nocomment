#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
commit=$(tr -d '[:space:]' < "$root/grammars/pinned-commit.txt")
dst="$root/grammars/golang"
mkdir -p "$dst"
base="https://raw.githubusercontent.com/antlr/grammars-v4/$commit/golang"
for name in GoLexer.g4 GoParser.g4; do
  curl -fsSL -o "$dst/$name" "$base/$name"
done
