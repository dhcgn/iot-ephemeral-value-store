[![Go](https://github.com/dhcgn/iot-ephemeral-value-store/actions/workflows/build_and_test.yml/badge.svg)](https://github.com/dhcgn/iot-ephemeral-value-store/actions/workflows/build_and_test.yml)
[![codecov](https://codecov.io/gh/dhcgn/iot-ephemeral-value-store/graph/badge.svg?token=LLTOitLLDc)](https://codecov.io/gh/dhcgn/iot-ephemeral-value-store)
[![Go Report Card](https://goreportcard.com/badge/github.com/dhcgn/iot-ephemeral-value-store)](https://goreportcard.com/report/github.com/dhcgn/iot-ephemeral-value-store)

# iot-ephemeral-value-store

A lightweight HTTP server for temporary IoT data storage. Upload sensor data with simple HTTP GET requests, retrieve it as JSON or plain text, and let it expire automatically. Perfect for IoT devices that can only make basic HTTP calls.

## What is this?

This server provides ephemeral (temporary) storage for IoT sensor data with a simple key-based system:
- **Upload data**: IoT devices use an upload key to send data via GET requests
- **Download data**: Applications use a separate download key to retrieve data
- **Automatic expiration**: Data is automatically deleted after a configurable period
- **No complex authentication**: Just keys - simple enough for any IoT device

The separation of upload and download keys means you can safely share read-only access to your data without exposing write capabilities.

## Key Features

- **Simple HTTP GET**: All operations work with GET requests - compatible with simple IoT devices
- **Key Pair System**: Separate upload (write) and download (read) keys for security
- **JSON & Plain Text**: Download data in JSON format or as individual plain text values
- **Patch Feature**: Combine multiple uploads into nested JSON structures
- **Auto Expiration**: Data automatically deleted after configured duration (e.g., 24h)
- **Web Viewer**: Built-in web interface to monitor keys in real-time
- **No Database Setup**: Embedded BadgerDB - just run the binary

## Quick Start Example

**1. Create a key pair:**
```bash
curl https://iot.hdev.io/kp
```
```json
{
  "upload-key": "abb0c54c...",
  "download-key": "233e9c1f..."
}
```

**2. Upload data (from IoT device):**
```bash
curl "https://iot.hdev.io/u/abb0c54c.../?tempC=23.5&humidity=45"
```

**3. Download data (from application):**
```bash
curl https://iot.hdev.io/d/233e9c1f.../json
```
```json
{
  "tempC": "23.5",
  "humidity": "45",
  "timestamp": "2024-12-29T18:51:08Z"
}
```

**Try it now!** A test installation is available at **[iot.hdev.io](https://iot.hdev.io)** for evaluation and fair use.

## Use Cases

### Shelly Devices
Shelly smart home devices can make HTTP GET requests. Upload sensor data from Shellys without complex integrations:
```bash
# Configure in Shelly: Actions → Webhook
https://iot.hdev.io/u/YOUR_UPLOAD_KEY/?temp={{temperature}}&power={{power}}
```

### Home Assistant
Upload sensor data from Home Assistant to share with external applications or mobile apps:
```yaml
rest_command:
  upload_sensor:
    url: "https://iot.hdev.io/u/{{ key }}/?value={{ value }}"
```
See **[README.HomeAssistant.md](README.HomeAssistant.md)** for complete integration guide.

### ESP32/ESP8266 Devices
Any device that can make HTTP GET requests:
```cpp
// ESP32/ESP8266 example
String url = "https://iot.hdev.io/u/YOUR_KEY/?temp=" + String(temp);
http.begin(url);
http.GET();
```

### Custom IoT Solutions
- DIY sensor networks
- Remote monitoring systems
- Data bridges between systems
- Temporary data exchange between services

## Installation

### Docker (Recommended)

**Basic usage:**
```bash
docker run -p 8080:8080 dhcgn/iot-ephemeral-value-store-server
```

**With persistent storage and custom retention:**
```bash
docker run -p 8080:8080 \
  -v /path/to/data:/data \
  dhcgn/iot-ephemeral-value-store-server \
  -persist-values-for 48h -store /data
```

**Docker Compose:**
```yaml
version: "3.3"
services:
  iot-server:
    image: dhcgn/iot-ephemeral-value-store-server
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
    command: -persist-values-for 24h -store /data
```

For Docker Compose with Traefik/HTTPS setup, see **[README.TechDetails.md](README.TechDetails.md#deployment-options)**.

### Run Binary
```bash
iot-ephemeral-value-store-server -persist-values-for 24h -port 8080
```

### Install as System Service (Linux)
```bash
sudo -i
bash <(curl -s https://raw.githubusercontent.com/dhcgn/iot-ephemeral-value-store/main/install/download-and-install.sh)
```

### Command Line Options
- `-persist-values-for`: Data retention duration (default: "24h")
- `-store`: Storage directory path (default: "./data")
- `-port`: Server port (default: 8080)

## Built-in Web Interface

Access your server at `http://localhost:8080/`:
- **Info Page** (`/`): Getting started guide, server stats, API examples
- **Viewer** (`/viewer`): Real-time monitoring tool for multiple keys

## Documentation

- **[README.TechDetails.md](README.TechDetails.md)** - Complete API reference, deployment options, technical details
- **[README.HomeAssistant.md](README.HomeAssistant.md)** - Home Assistant integration guide
- **[CLAUDE.md](CLAUDE.md)** - Development guide for contributors

## API Overview

| Operation | Endpoint | Description |
|-----------|----------|-------------|
| Create key pair | `GET /kp` | Generate upload/download key pair |
| Upload data | `GET /u/{uploadKey}?param=value` | Upload/replace data |
| Patch data | `GET /patch/{uploadKey}/path?param=value` | Merge data into nested structure |
| Download JSON | `GET /d/{downloadKey}/json` | Get all data as JSON |
| Download plain | `GET /d/{downloadKey}/plain/{param}` | Get single value as plain text |
| Delete data | `GET /delete/{uploadKey}` | Delete all data for this key |

See **[README.TechDetails.md](README.TechDetails.md)** for complete API documentation.

## Why Use This?

- **No authentication complexity**: Just keys - perfect for simple IoT devices
- **Privacy by design**: Separate upload/download keys prevent unauthorized writes
- **Zero configuration**: No database setup, no complex configs
- **Automatic cleanup**: Data expires automatically - no maintenance needed
- **Simple integration**: Works with any device that can make HTTP GET requests
- **Self-hosted**: Run on your network, keep your data private

## Example: Patch Feature

Build complex data structures from multiple uploads:

```bash
# Upload from different sensors to the same key
curl "https://iot.hdev.io/patch/YOUR_KEY/living_room/?temp=22"
curl "https://iot.hdev.io/patch/YOUR_KEY/bedroom/?temp=20"
curl "https://iot.hdev.io/patch/YOUR_KEY/basement/?temp=18"
```

Download once to get everything:
```bash
curl https://iot.hdev.io/d/YOUR_DOWNLOAD_KEY/json
```
```json
{
  "living_room": {
    "temp": "22",
    "timestamp": "2024-12-29T18:51:07Z"
  },
  "bedroom": {
    "temp": "20",
    "timestamp": "2024-12-29T18:51:08Z"
  },
  "basement": {
    "temp": "18",
    "timestamp": "2024-12-29T18:51:09Z"
  },
  "timestamp": "2024-12-29T18:51:09Z"
}
```

## License

Licensed under the MIT License. See [LICENSE](LICENSE) file for details.

## Links

- **GitHub**: [github.com/dhcgn/iot-ephemeral-value-store](https://github.com/dhcgn/iot-ephemeral-value-store)
- **Docker Hub**: [dhcgn/iot-ephemeral-value-store-server](https://hub.docker.com/r/dhcgn/iot-ephemeral-value-store-server)
- **Test Installation**: [iot.hdev.io](https://iot.hdev.io) (fair use)
