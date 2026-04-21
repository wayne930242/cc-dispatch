#!/usr/bin/env sh
# Install cc-dispatch (ccd) from the latest GitHub Release.
set -eu

REPO="wayne930242/cc-dispatch"
INSTALL_DIR="${CC_DISPATCH_INSTALL_DIR:-$HOME/.cc-dispatch/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac
case "$os" in
  linux|darwin) ;;
  *) echo "unsupported OS: $os (Windows: download the zip manually from GitHub Releases)" >&2; exit 1 ;;
esac

version="${CC_DISPATCH_VERSION:-$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')}"
if [ -z "$version" ]; then
  echo "could not resolve latest version" >&2
  exit 1
fi

filename="cc-dispatch_${version#v}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/$version/$filename"
checksums_url="https://github.com/$REPO/releases/download/$version/checksums.txt"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "→ downloading $filename"
curl -fsSL "$url" -o "$tmp/$filename"

echo "→ verifying checksum"
curl -fsSL "$checksums_url" -o "$tmp/checksums.txt"
(cd "$tmp" && grep " $filename\$" checksums.txt | shasum -a 256 -c - 2>/dev/null || grep " $filename\$" checksums.txt | sha256sum -c -)

echo "→ extracting"
tar -xzf "$tmp/$filename" -C "$tmp"

mkdir -p "$INSTALL_DIR"
install -m 0755 "$tmp/ccd" "$INSTALL_DIR/ccd"

echo
echo "✓ installed: $INSTALL_DIR/ccd ($version)"

# Try to symlink into a directory already on PATH so fresh shells and
# headless MCP spawns can find ccd without shell-profile edits.
linked=""
case ":$PATH:" in
  *":$INSTALL_DIR:"*) linked="already-on-path" ;;
  *)
    for pathdir in "$HOME/.local/bin" "/usr/local/bin"; do
      if [ -d "$pathdir" ] && [ -w "$pathdir" ]; then
        ln -sf "$INSTALL_DIR/ccd" "$pathdir/ccd"
        echo "→ linked: $pathdir/ccd -> $INSTALL_DIR/ccd"
        linked="symlinked"
        break
      fi
    done
    ;;
esac

if [ -z "$linked" ]; then
  echo
  echo "Add to your shell profile:"
  echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
fi
