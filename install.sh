#!/usr/bin/env bash
set -e

REPO="warmdev17/ash"
BIN="ash"
INSTALL_DIR="$HOME/.local/bin"

OS="linux"
ARCH=$(uname -m)

case "$ARCH" in
x86_64) ARCH="amd64" ;;
aarch64) ARCH="arm64" ;;
*)
  echo "❌ Unsupported architecture: $ARCH"
  exit 1
  ;;
esac

TAG=$(curl -fsSL https://api.github.com/repos/$REPO/releases/latest |
  grep '"tag_name"' | cut -d '"' -f 4)

if [ -z "$TAG" ]; then
  echo "❌ Failed to detect latest release"
  exit 1
fi

ARCHIVE="${BIN}_${TAG#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$TAG/$ARCHIVE"

echo "⬇️  Downloading $ARCHIVE"
curl -fL "$URL" -o "$ARCHIVE"

tar -xzf "$ARCHIVE"
chmod +x "$BIN"

mkdir -p "$INSTALL_DIR"
mv "$BIN" "$INSTALL_DIR/"

rm "$ARCHIVE"

echo "✅ ash installed to $INSTALL_DIR/$BIN"

if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo
  echo "⚠️  $INSTALL_DIR is not in your PATH"
  echo "Add this to your shell config:"
  echo 'export PATH="$HOME/.local/bin:$PATH"'
fi
