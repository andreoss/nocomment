#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
tracker=${1:-doc/Tracker.adoc}

if [ ! -f "$tracker" ]; then
  echo "missing $tracker" >&2
  exit 1
fi

status=0
claimed=0
for pair in $(sed -n 's/^| \(ID-[0-9]*\) | M[0-9]* | \([-x!o~]\) | \(ID-[0-9]*\) | M[0-9]* | \([-x!o~]\).*/\1=\2 \3=\4/p' "$tracker"); do
  id=${pair%=*}
  state=${pair#*=}
  case $state in
  x | !) ;;
  *) continue ;;
  esac
  claimed=$((claimed + 1))
  if ! grep -rql "$id" doc/review/; then
    echo "$id is marked $state and no sprint review names it" >&2
    status=1
  fi
  if ! grep -ql "$id" doc/Pilot.adoc; then
    echo "$id is marked $state and the backlog has no item text" >&2
    status=1
  fi
done

if [ "$claimed" -eq 0 ]; then
  echo "no claimed id found in $tracker" >&2
  exit 1
fi
if [ "$status" -ne 0 ]; then
  echo "evidence audit: $tracker breaches R-23" >&2
  exit 1
fi
echo "evidence audit: $claimed claimed ids are named by a review and the backlog"
