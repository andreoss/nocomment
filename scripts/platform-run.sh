#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"

version=v7.2.0-1
digest=dce64b2dc6b005485c7aa735a7ea39cb0006bf7e5badc28b324b2cd0c73d883f
emulator=scratch/tools/qemu-aarch64-static
url=https://github.com/multiarch/qemu-user-static/releases/download/$version/qemu-aarch64-static

mkdir -p scratch/tools
if [ ! -x "$emulator" ]; then
  if ! curl -sSL -m 180 -H 'Connection: close' -o "$emulator" "$url"; then
    echo "emulator unavailable: $url" >&2
    exit 1
  fi
  chmod +x "$emulator"
fi

have=$(sha256sum "$emulator" | awk '{print $1}')
if [ "$have" != "$digest" ]; then
  echo "emulator digest $have, expected $digest" >&2
  exit 1
fi

echo "commit: $(git rev-parse HEAD)"
echo "emulator: qemu-aarch64 $version"
GOOS=linux GOARCH=arm64 go test -exec="$root/$emulator" ./...
