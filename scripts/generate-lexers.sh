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
rm -rf internal/lexer/generated
mkdir -p internal/lexer/generated
tab=$(printf '\t')
while IFS="$tab" read -r lang exts pkg dir main extra base transform patch; do
  case "$lang" in
    lang|"") continue ;;
  esac
  work="scratch/grammars/$lang"
  rm -rf "$work"
  mkdir -p "$work"
  cp "grammars/$lang/$main" "$work/$main"
  case "$patch" in
    skip)
      sed 's/-> skip/-> channel(HIDDEN)/g' "$work/$main" > "$work/patched"
      mv "$work/patched" "$work/$main"
      ;;
    python)
      awk '
        {
          if (index($0, "( SPACES | COMMENT | LINE_JOINING)")) {
            sub(/\( SPACES \| COMMENT \| LINE_JOINING\)/, "( SPACES | LINE_JOINING)")
          }
          if (index($0, "fragment COMMENT:")) {
            sub(/fragment COMMENT:/, "COMMENT:")
            sub(/;[ \t]*$/, " -> channel(HIDDEN);")
          }
          print
        }
      ' "$work/$main" > "$work/patched"
      mv "$work/patched" "$work/$main"
      ;;
  esac
  case "$transform" in
    this)
      awk '
        {
          if (index($0, "this.") && index($0, "}?")) {
            gsub(/this\./, "p.")
          } else if (index($0, "this.")) {
            gsub(/this\./, "l.")
          }
          print
        }
      ' "$work/$main" > "$work/patched"
      mv "$work/patched" "$work/$main"
      ;;
  esac
  out="internal/lexer/generated/$pkg"
  mkdir -p "$out"
  java -jar "$jar" -Dlanguage=Go -package "$pkg" -Xexact-output-dir -lib "grammars/$lang" -o "$out" "$work/$main"
  if [ "$base" != "-" ] && [ -n "$base" ]; then
    oldifs=$IFS
    IFS=,
    for f in $base; do
      sed -E "s/^package parser$/package $pkg/" "grammars/$lang/$f.txt" > "$out/$f"
    done
    IFS=$oldifs
  fi
done < grammars/languages.tsv
