#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
version=$(tr -d '[:space:]' < "$root/tools/antlr-version.txt")
jar="$root/tools/antlr-$version-complete.jar"
if [ ! -f "$jar" ]; then
  mkdir -p "$root/tools"
  curl -fsSL -o "$jar" "https://repo1.maven.org/maven2/org/antlr/antlr4/$version/antlr4-$version-complete.jar"
fi
cd "$root"
out="internal/lexer/generated/golang"
rm -rf "$out"
mkdir -p "$out"
java -jar "$jar" -Dlanguage=Go -package golang -Xexact-output-dir -o "$out" "grammars/golang/GoLexer.g4"
