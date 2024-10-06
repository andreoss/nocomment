#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
bin="$root/bin/nocomment"
if [ ! -x "$bin" ]; then
  echo "build bin/nocomment first" >&2
  exit 1
fi
mkdir -p "$root/scratch"
file="$root/scratch/measure.go"
printf 'package main // measure\n// c\nvar x = 1\n' > "$file"
runs=20
measure() {
  total=0
  i=0
  while [ "$i" -lt "$runs" ]; do
    start=$(date +%s%N)
    "$@" >/dev/null 2>&1
    end=$(date +%s%N)
    total=$((total + (end - start) / 1000))
    i=$((i + 1))
  done
  echo $((total / runs))
}
version_us=$(measure "$bin" -version)
file_us=$(measure "$bin" -l "$file")
rm -f "$file"
echo "version startup: ${version_us} us"
echo "one file: ${file_us} us"
