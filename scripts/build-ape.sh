#!/bin/sh
# Build crm.com as a Cosmopolitan APE that embeds the host Go binary.
# Implements: SYS-REQ-260821-AFPN SW-REQ-260821-AC3S
set -eu

ROOT=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
COSMOCC_HOME=${COSMOCC_HOME:-"$HOME/.local/cosmocc"}
export PATH="$COSMOCC_HOME/bin:$HOME/.local/bin:$PATH"

# cosmocc.zip extract often leaves cc1/as/ld without +x.
if [ -d "$COSMOCC_HOME/libexec" ]; then
  chmod -R +x "$COSMOCC_HOME/bin" "$COSMOCC_HOME/libexec" 2>/dev/null || true
fi

if ! command -v cosmocc >/dev/null 2>&1; then
  echo "cosmocc not found. Install from https://cosmo.zip/pub/cosmocc/cosmocc.zip into $COSMOCC_HOME" >&2
  exit 1
fi

OUT=${1:-"$ROOT/crm.com"}
HOST_OS=$(uname -s | tr 'A-Z' 'a-z')
HOST_ARCH=$(uname -m)
case "$HOST_ARCH" in
  arm64|aarch64) HOST_ARCH=arm64 ;;
  x86_64|amd64) HOST_ARCH=x86_64 ;;
esac
BLOB_NAME="crm.${HOST_OS}-${HOST_ARCH}"

BUILDDIR="$ROOT/.build/ape"
mkdir -p "$BUILDDIR"
echo "building native go binary"
(cd "$ROOT" && go build -o "$BUILDDIR/$BLOB_NAME" ./cmd/crm)
cp "$BUILDDIR/$BLOB_NAME" "$BUILDDIR/crm"

HELLO="$BUILDDIR/hello.com"
echo "building hello.com"
cosmocc -o "$HELLO" "$ROOT/ape/hello.c"

echo "building wrapper APE"
cosmocc -o "$OUT" "$ROOT/ape/wrapper.c"

if command -v zip >/dev/null 2>&1; then
  (cd "$BUILDDIR" && zip -q -A "$OUT" "$BLOB_NAME" crm)
else
  echo "zip not found; APE wrapper has no embedded blob" >&2
  exit 1
fi

if [ "$(uname -m)" = "arm64" ] && [ ! -x "$HOME/.local/bin/ape" ]; then
  APE_C=$(find "$COSMOCC_HOME" -name 'ape-m1.c' 2>/dev/null | head -n 1)
  if [ -n "$APE_C" ]; then
    mkdir -p "$HOME/.local/bin"
    cc -O -o "$HOME/.local/bin/ape" "$APE_C"
    echo "built Apple Silicon ape loader at $HOME/.local/bin/ape"
  fi
fi

echo "APE written to $OUT"
echo "try: chmod +x $OUT && $OUT --help"
