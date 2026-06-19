#!/usr/bin/env bash
set -euo pipefail

# SpecRalph installer script
# Downloads and installs the latest SpecRalph release binary

REPO="iyaki/specralph"
BIN_NAME="ralph"

# Detect platform
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64|amd64) ARCH_GO="amd64" ;;
  aarch64|arm64) ARCH_GO="arm64" ;;
  *)
    echo "Error: Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  linux) OS_GO="linux" ;;
  darwin) OS_GO="darwin" ;;
  msys*|mingw*|cygwin*) OS_GO="windows" ;;
  *)
    echo "Error: Unsupported OS: $OS"
    exit 1
    ;;
esac

# Get the latest release version from GitHub
LATEST_TAG="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep -o '"tag_name": "v[^"]*' | cut -d'"' -f4)"
if [ -z "$LATEST_TAG" ]; then
  echo "Error: Could not fetch latest release version"
  exit 1
fi

VERSION="${LATEST_TAG#v}"
echo "Installing SpecRalph $VERSION for $OS_GO-$ARCH_GO..."

# Build the download URL
BINARY_NAME="ralph_v${VERSION}_${OS_GO}_${ARCH_GO}"
if [ "$OS_GO" = "windows" ]; then
  BINARY_NAME="${BINARY_NAME}.exe"
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/${LATEST_TAG}/${BINARY_NAME}"
INSTALL_DIR="/usr/local/bin"
INSTALL_PATH="${INSTALL_DIR}/${BIN_NAME}"

# Check if we can write to /usr/local/bin, otherwise use ~/.local/bin
if [ ! -w "$INSTALL_DIR" ]; then
  INSTALL_DIR="$HOME/.local/bin"
  INSTALL_PATH="${INSTALL_DIR}/${BIN_NAME}"
  mkdir -p "$INSTALL_DIR"
fi

# Download and install
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading from $DOWNLOAD_URL..."
curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/${BIN_NAME}"

# Make executable
chmod +x "${TMP_DIR}/${BIN_NAME}"

# Install
mv "${TMP_DIR}/${BIN_NAME}" "$INSTALL_PATH"

echo ""
echo "✓ SpecRalph installed successfully to $INSTALL_PATH"
echo ""
echo "Verify installation:"
echo "  ralph --help"
echo "  ralph version"
echo ""

# Check if installation directory is in PATH
if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
  echo "⚠ Warning: $INSTALL_DIR is not in your PATH"
  echo ""
  if [ "$INSTALL_DIR" = "$HOME/.local/bin" ]; then
    echo "Add this to your shell profile:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
  else
    echo "You may need to add $INSTALL_DIR to your PATH"
  fi
fi