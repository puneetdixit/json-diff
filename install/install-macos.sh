#!/bin/bash
set -e

BINARY_URL="https://github.com/puneetdixit/json-diff/releases/latest/download/json-diff-darwin-amd64"
INSTALL_PATH="/usr/local/bin/json-diff"

echo "Downloading json-diff for macOS..."
curl -L "$BINARY_URL" -o "$INSTALL_PATH"
chmod +x "$INSTALL_PATH"
echo "Installed json-diff to $INSTALL_PATH"
echo "You can now run 'json-diff' from your terminal"
