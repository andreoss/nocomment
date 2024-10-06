#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
base_file=doc/ops/commit-policy-base.txt

check() {
  label=$1
  message=$2
  ids=$(printf '%s\n' "$message" | grep -oE 'ID-[0-9]+' | sort -u || true)
  found=0
  if [ -n "$ids" ]; then
    found=$(printf '%s\n' "$ids" | wc -l | tr -d ' ')
  fi
  if [ "$found" -ne 1 ]; then
    echo "$label: expected one backlog id, found $found" >&2
    return 1
  fi
  subject=$(printf '%s\n' "$message" | sed -n '1p')
  case $subject in
  *"$ids"*) ;;
  *)
    echo "$label: id $ids is absent from the subject" >&2
    return 1
    ;;
  esac
}

if [ "${1:-}" = "-" ]; then
  if check message "$(cat)"; then
    echo "commit policy: one backlog id"
    exit 0
  fi
  exit 1
fi

range=${1:-}
if [ -z "$range" ]; then
  if [ ! -f "$base_file" ]; then
    echo "missing $base_file" >&2
    exit 1
  fi
  range="$(sed -n '1p' "$base_file" | tr -d '[:space:]')..HEAD"
fi

status=0
count=0
for sha in $(git rev-list "$range"); do
  count=$((count + 1))
  if ! check "$(git log -1 --format=%h "$sha")" "$(git log -1 --format=%B "$sha")"; then
    status=1
  fi
done
if [ "$status" -ne 0 ]; then
  echo "commit policy: $range violates R-25" >&2
  exit 1
fi
echo "commit policy: $count commits in $range carry one id each"
