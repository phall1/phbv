#!/bin/sh
# phbv installer — curl -fsSL https://raw.githubusercontent.com/phall1/phbv/main/install.sh | sh
#
# Downloads the latest release binary for your OS/arch from GitHub Releases and
# installs it. Override the install dir with PREFIX, e.g.
#   curl -fsSL .../install.sh | PREFIX=/usr/local/bin sh
set -eu

REPO="phall1/phbv"
BIN="phbv"
PREFIX="${PREFIX:-$HOME/.local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  *) echo "phbv: unsupported architecture: $arch" >&2; exit 1 ;;
esac
case "$os" in
  linux | darwin) ;;
  *) echo "phbv: unsupported OS: $os (use 'go install $REPO@latest')" >&2; exit 1 ;;
esac

echo "phbv: resolving latest release…"
tag=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' | head -1 | cut -d'"' -f4)
if [ -z "$tag" ]; then
  echo "phbv: could not find a published release. Try: go install $REPO@latest" >&2
  exit 1
fi

# Must match .goreleaser.yaml archives.name_template
asset="${BIN}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/$tag/$asset"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
echo "phbv: downloading $tag ($asset)…"
curl -fsSL "$url" -o "$tmp/$asset"
tar -xzf "$tmp/$asset" -C "$tmp"

mkdir -p "$PREFIX"
install -m 0755 "$tmp/$BIN" "$PREFIX/$BIN"
echo "phbv: installed $tag to $PREFIX/$BIN"

case ":$PATH:" in
  *":$PREFIX:"*) ;;
  *) echo "phbv: note — $PREFIX is not on your PATH. Add it to your shell profile." ;;
esac
