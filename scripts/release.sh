#!/usr/bin/env sh
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$root"
version=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo devel)}
dist="$root/dist"
rm -rf "$dist"
mkdir -p "$dist"
targets="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64"
for target in $targets; do
  os=${target%/*}
  arch=${target#*/}
  name="nocomment_${version}_${os}_${arch}"
  work="$dist/$name"
  mkdir -p "$work"
  binary="nocomment"
  if [ "$os" = "windows" ]; then
    binary="nocomment.exe"
  fi
  GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w -X nocomment/internal/cli.Version=$version" \
    -o "$work/$binary" ./cmd/nocomment
  if [ "$os" = "windows" ]; then
    (cd "$dist" && zip -q -r "$name.zip" "$name")
  else
    (cd "$dist" && tar -czf "$name.tar.gz" "$name")
  fi
  rm -rf "$work"
done
(cd "$dist" && sha256sum ./*.tar.gz ./*.zip > SHA256SUMS)
echo "release $version"
ls -1 "$dist"
