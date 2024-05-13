#!/bin/bash

# Define the service name and default store path
SERVICE_NAME="iot-ephemeral-value-store-server"
USER="root"
GROUP="root"
DEFAULT_STORE_PATH="/var/lib/iot-ephemeral-value-store"

# Check if running as root
if [ "$(id -u)" != "0" ]; then
    echo "This script must be run as root" 1>&2
    exit 1
fi

# Check if binary path is provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 /path/to/binary"
    exit 1
fi

BINARY_PATH=$1

# Validate that the binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary not found at $BINARY_PATH"
    exit 1
fi

# Stop the service if it's running
if systemctl is-active --quiet $SERVICE_NAME; then
    echo "Stopping service..."
    systemctl stop $SERVICE_NAME
fi

# Copy the binary to /usr/local/bin
INSTALL_PATH="/usr/local/bin/$SERVICE_NAME"
echo "Copying binary to $INSTALL_PATH..."
cp "$BINARY_PATH" $INSTALL_PATH
chmod +x $INSTALL_PATH

# Ensure the default store directory exists
if [ ! -d "$DEFAULT_STORE_PATH" ]; then
    echo "Creating store directory at $DEFAULT_STORE_PATH..."
    mkdir -p $DEFAULT_STORE_PATH
    chown $USER:$GROUP $DEFAULT_STORE_PATH
fi

# Create systemd service file
echo "Creating systemd service file..."
cat > /etc/systemd/system/$SERVICE_NAME.service <<EOF
[Unit]
Description=IoT Ephemeral Value Store Server
After=network.target

[Service]
Type=simple
User=$USER
Group=$GROUP
ExecStart=$INSTALL_PATH -store $DEFAULT_STORE_PATH
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd, enable and start service
echo "Reloading systemd daemon and enabling service..."
systemctl daemon-reload
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME
systemctl status $SERVICE_NAME