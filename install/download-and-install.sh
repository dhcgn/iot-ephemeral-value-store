#!/bin/bash

# Variables
GITHUB_REPO="dhcgn/iot-ephemeral-value-store"
ASSET_NAME="iot-ephemeral-value-store-server-linux-am64.bin"
INSTALL_SCRIPT="install-service.sh"
TEMP_DIR=$(mktemp -d)

# Check if port parameter is provided
PORT=""
if [ "$#" -eq 1 ]; then
    if ! [[ $1 =~ ^[0-9]+$ ]] || [ $1 -lt 1 ] || [ $1 -gt 65535 ]; then
        echo "Error: Invalid port number"
        exit 1
    fi
    PORT=$1
fi

# Function to clean up temporary directory
cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# 1. Download latest release asset and print the version number
echo "Fetching latest release information..."
LATEST_RELEASE_URL=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep "browser_download_url.*$ASSET_NAME" | cut -d '"' -f 4)

# print lastest release version number
LATEST_RELEASE_VERSION=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep "tag_name" | cut -d '"' -f 4)
echo "Latest release version: $LATEST_RELEASE_VERSION"

if [ -z "$LATEST_RELEASE_URL" ]; then
    echo "Error: Unable to find the latest release asset."
    exit 1
fi

echo "Downloading latest release asset..."
curl -L "$LATEST_RELEASE_URL" -o "$TEMP_DIR/$ASSET_NAME"
chmod +x "$TEMP_DIR/$ASSET_NAME"

# 2. Download 'install-service.sh'
INSTALL_SCRIPT_URL="https://raw.githubusercontent.com/$GITHUB_REPO/main/install/$INSTALL_SCRIPT"

echo "Downloading install-service.sh..."
curl -L "$INSTALL_SCRIPT_URL" -o "$TEMP_DIR/$INSTALL_SCRIPT"
chmod +x "$TEMP_DIR/$INSTALL_SCRIPT"

# 3. Run 'install-service.sh' with the binary path and optional port
echo "Running install-service.sh with the binary path..."
if [ -n "$PORT" ]; then
    "$TEMP_DIR/$INSTALL_SCRIPT" "$TEMP_DIR/$ASSET_NAME" "$PORT"
else
    "$TEMP_DIR/$INSTALL_SCRIPT" "$TEMP_DIR/$ASSET_NAME"
fi

echo "Done."