#!/bin/bash

# Variables
GITHUB_REPO="dhcgn/iot-ephemeral-value-store" # Replace with the actual GitHub repo
ASSET_NAME="iot-ephemeral-value-store-server.tar.gz"
INSTALL_SCRIPT="install-service.sh"
TEMP_DIR=$(mktemp -d)

# Function to clean up temporary directory
cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

# 1. Download latest release asset 'iot-ephemeral-value-store-server.tar.gz' from GitHub
echo "Fetching latest release information..."
LATEST_RELEASE_URL=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep "browser_download_url.*$ASSET_NAME" | cut -d '"' -f 4)

if [ -z "$LATEST_RELEASE_URL" ]; then
    echo "Error: Unable to find the latest release asset."
    exit 1
fi

echo "Downloading latest release asset..."
curl -L "$LATEST_RELEASE_URL" -o "$TEMP_DIR/$ASSET_NAME"

# 2. Download 'install-service.sh'
INSTALL_SCRIPT_URL="https://raw.githubusercontent.com/$GITHUB_REPO/main/install/$INSTALL_SCRIPT"

echo "Downloading install-service.sh..."
curl -L "$INSTALL_SCRIPT_URL" -o "$TEMP_DIR/$INSTALL_SCRIPT"
chmod +x "$TEMP_DIR/$INSTALL_SCRIPT"

# 3. Extract 'iot-ephemeral-value-store-server.tar.gz' and run 'install-service.sh' with the path to the binary
echo "Extracting the tar.gz file..."
tar -xzvf "$TEMP_DIR/$ASSET_NAME" -C "$TEMP_DIR"

BINARY_PATH=$(find "$TEMP_DIR" -type f -name 'iot-ephemeral-value-store-server')

if [ -z "$BINARY_PATH" ]; then
    echo "Error: Unable to find the binary file."
    exit 1
fi

echo "Running install-service.sh with the binary path..."
"$TEMP_DIR/$INSTALL_SCRIPT" "$BINARY_PATH"

echo "Done."