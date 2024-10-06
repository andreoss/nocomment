#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
tracker=${1:-doc/Tracker.adoc}
sprints=doc/Sprints.adoc

if [ ! -f "$tracker" ]; then
  echo "missing $tracker" >&2
  exit 1
fi

rows() {
  awk -F'|' '
    /^\| *ID-[0-9]+ *\|/ {
      gsub(/ /, "")
      if ($2 ~ /^ID-[0-9]+$/ && $4 ~ /^[-x!o~]$/) printf "%s=%s\n", $2, $4
      if ($5 ~ /^ID-[0-9]+$/ && $7 ~ /^[-x!o~]$/) printf "%s=%s\n", $5, $7
    }
  ' "$1"
}

owner_of() {
  awk -v id="$1" '
    /^== S/ { sprint = $2 }
    /^- ids:/ {
      n = split($0, part, /[ ,:]+/)
      for (i = 1; i <= n; i++) {
        if (part[i] == id) {
          print sprint
          exit
        }
      }
    }
  ' "$sprints"
}

status=0
claimed=0
for pair in $(rows "$tracker"); do
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
  if [ "$state" = "!" ]; then
    owner=$(owner_of "$id")
    if [ -n "$owner" ] && [ -f "doc/review/$owner.adoc" ] &&
      grep -q '^Status: done$' "doc/review/$owner.adoc"; then
      echo "$id is blocked and $owner claims done" >&2
      status=1
    fi
  fi
done

reviewed=0
for sprint in $(sed -n 's/^== \(S[0-9][0-9]*\).*/\1/p' "$sprints"); do
  review=doc/review/$sprint.adoc
  if [ ! -f "$review" ]; then
    echo "$sprint has no review" >&2
    status=1
    continue
  fi
  if ! grep -qE '^Status: (done|blocked)$' "$review"; then
    echo "$review carries no status" >&2
    status=1
    continue
  fi
  reviewed=$((reviewed + 1))
done

if [ "$claimed" -eq 0 ]; then
  echo "no claimed id found in $tracker" >&2
  exit 1
fi
if [ "$status" -ne 0 ]; then
  echo "evidence audit: $tracker breaches R-23" >&2
  exit 1
fi
echo "evidence audit: $claimed claimed ids evidenced, $reviewed sprints reviewed"
