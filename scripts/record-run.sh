#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"

usage() {
  echo "usage: record-run.sh <sprint> <target> <identifier> [result]" >&2
  echo "  sprint     S33 or S47" >&2
  echo "  target     windows/amd64, darwin/amd64, darwin/arm64, linux/arm64" >&2
  echo "  identifier run URL or run id that anyone can open" >&2
  echo "  result     green or red, default green" >&2
  exit 2
}

[ $# -ge 3 ] || usage
sprint=$1
target=$2
identifier=$3
result=${4:-green}

case $sprint in
S33 | S47) ;;
*) usage ;;
esac
case $target in
windows/amd64 | darwin/amd64 | darwin/arm64 | linux/amd64 | linux/arm64) ;;
*) usage ;;
esac
case $result in
green | red) ;;
*) usage ;;
esac

record=doc/ops/records/$sprint.adoc
if [ ! -f "$record" ]; then
  echo "missing $record" >&2
  exit 1
fi

commit=$(git rev-parse HEAD)
{
  echo
  echo "== Hosted run: $target"
  echo "Commit under test: $commit"
  echo "Run: $identifier"
  echo "Result: $result"
  echo
  echo "Recorded by \`scripts/record-run.sh\`. R-27 is satisfied for this"
  echo "target when the run above is open to a reader and reports $result."
} >>"$record"

echo "appended $target to $record"
echo "next: set the id's status in doc/Tracker.adoc once every target is recorded"
