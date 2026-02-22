[![Go](https://github.com/dhcgn/iot-ephemeral-value-store/actions/workflows/build_and_test.yml/badge.svg)](https://github.com/dhcgn/iot-ephemeral-value-store/actions/workflows/build_and_test.yml)
[![codecov](https://codecov.io/gh/dhcgn/iot-ephemeral-value-store/graph/badge.svg?token=LLTOitLLDc)](https://codecov.io/gh/dhcgn/iot-ephemeral-value-store)
[![Go Report Card](https://goreportcard.com/badge/github.com/dhcgn/iot-ephemeral-value-store)](https://goreportcard.com/report/github.com/dhcgn/iot-ephemeral-value-store)
[![govulncheck](https://github.com/dhcgn/iot-ephemeral-value-store/actions/workflows/govulncheck.yml/badge.svg)](https://github.com/dhcgn/iot-ephemeral-value-store/actions/workflows/govulncheck.yml)

> MCP Demo Server
> Url: <https://iot.hdev.io/mcp>  
> [![Install in VS Code](https://img.shields.io/badge/Install_in-VS_Code-0098FF?style=flat-square&logo=visualstudiocode&logoColor=white)](https://vscode.dev/redirect/mcp/install?name=iot-ephemeral-value-store&config=%7B%22type%22%3A%22http%22%2C%22url%22%3A%22%22%7D)  
> [![Install in VS Code Insiders](https://img.shields.io/badge/Install_in-VS_Code_Insiders-24bfa5?style=flat-square&logo=visualstudiocode&logoColor=white)](https://insiders.vscode.dev/redirect/mcp/install?name=iot-ephemeral-value-store&config=%7B%22type%22%3A%22http%22%2C%22url%22%3A%22%22%7D&quality=insiders)  
> [![Install in Visual Studio](https://img.shields.io/badge/Install_in-Visual_Studio-C16FDE?style=flat-square&logo=visualstudio&logoColor=white)](https://vs-open.link/mcp-install?%7B%22type%22%3A%22http%22%2C%22url%22%3A%22%22%7D)  
> [![Install in Cursor](https://img.shields.io/badge/Install_in-Cursor-000000?style=flat-square&logoColor=white)](https://cursor.com/en/install-mcp?name=iot-ephemeral-value-store&config=eyJ0eXBlIjoiaHR0cCIsInVybCI6IiJ9)

# iot-ephemeral-value-store

A lightweight HTTP server for temporary IoT data storage. Upload sensor data with simple HTTP GET requests, retrieve it as JSON or plain text, and let it expire automatically. Perfect for IoT devices that can only make basic HTTP calls.

Url: <https://iot.hdev.io/>

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
- **Optional Key Prefixes**: Upload keys have `u_` prefix, download keys have `d_` prefix for easy identification
- **JSON & Plain Text**: Download data in JSON format or as individual plain text values
- **Patch Feature**: Combine multiple uploads into nested JSON structures
- **Auto Expiration**: Data automatically deleted after configured duration (e.g., 24h)
- **Web Viewer**: Built-in web interface to monitor keys in real-time
- **No Database Setup**: Embedded BadgerDB - just run the binary

## Understanding Keys

### Key Prefixes (Optional)

To help distinguish between upload and download keys, the system adds optional prefixes:
- **Upload keys**: Start with `u_` (e.g., `u_f0cc4756b9b01bc9...`)
- **Download keys**: Start with `d_` (e.g., `d_2647b0c1b03c6885...`)

These prefixes are **completely optional**. You can use either format:
- **With prefix**: `https://iot.hdev.io/u/u_f0cc4756.../`
- **Without prefix**: `https://iot.hdev.io/u/f0cc4756.../`

Both formats are equivalent and work identically. The prefix is purely for human readability and doesn't affect functionality.

### Key Generation

When you request a new key pair from `/kp`, you receive both keys with prefixes:
```json
{
  "upload-key": "u_f0cc4756b9b01bc942866c171fb6b49113165a9740eae38e4525bd56782b5366",
  "download-key": "d_2647b0c1b03c6885daa32f0e6d231c5a21d321ed737a2a20aa96b545c093dbed"
}
```

The download key is cryptographically derived from the upload key using SHA256, ensuring:
- You cannot derive the upload key from a download key (one-way)
- The same upload key always produces the same download key
- Upload and download operations are securely separated

## Quick Start Example

**1. Create a key pair:**
```bash
curl https://iot.hdev.io/kp
```
```json
{
  "upload-key": "u_abb0c54c...",
  "download-key": "d_233e9c1f..."
}
```

**2. Upload data (from IoT device):**
```bash
# With prefix (recommended for clarity)
curl "https://iot.hdev.io/u/u_abb0c54c.../?tempC=23.5&humidity=45"

# Or without prefix (backward compatible)
curl "https://iot.hdev.io/u/abb0c54c.../?tempC=23.5&humidity=45"
```

**3. Download data (from application):**
```bash
# With prefix
curl https://iot.hdev.io/d/d_233e9c1f.../json

# Or without prefix
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

## Model Context Protocol (MCP) Integration

The server provides a **Model Context Protocol (MCP)** endpoint at `/mcp` for use with AI assistants and LLM-powered applications. MCP enables AI systems to directly interact with your IoT data store using standardized tools.

### Available MCP Tools

The MCP server exposes the following tools:

1. **generate_key_pair** - Generate a new upload/download key pair for secure data storage
2. **upload_data** - Upload data to the store (replaces existing data)
3. **patch_data** - Merge data into nested structures (preserves existing data)
4. **download_data** - Retrieve data by download key (supports full JSON or specific fields)
5. **delete_data** - Delete all data associated with an upload key

### Using MCP with Claude

To use the IoT Ephemeral Value Store with Claude or other MCP-compatible AI assistants:

**1. Get MCP Server Information:**
```bash
curl http://localhost:8080/mcp
```

**2. Configure in your MCP client:**

For Claude Desktop, add to `claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "iot-ephemeral-value-store": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

### Example MCP Usage

With MCP enabled, you can ask your AI assistant:

```
"Generate a key pair for my IoT device"
↓
MCP calls: generate_key_pair
↓
Assistant receives: {"upload_key": "...", "download_key": "..."}

"Store temperature readings of 23.5°C and humidity 45%"
↓
MCP calls: upload_data with received upload_key
↓
Assistant confirms: Data uploaded successfully

"What sensor readings do we have?"
↓
MCP calls: download_data with received download_key
↓
Assistant reads: {"tempC": "23.5", "humidity": "45", "timestamp": "..."}
```

### MCP Endpoint Details

- **Endpoint**: `GET /mcp` (for server information) or `POST /mcp` (for tool calls)
- **Protocol**: HTTP Streamable (SSE - Server-Sent Events)
- **Response Format**: JSON
- **Tools Available**: Dynamically retrieved from the server

### Benefits of MCP Integration

- **AI-Powered Workflows**: Let AI assistants manage your IoT data automatically
- **Natural Language Interface**: Describe what you want in plain English
- **Standardized Protocol**: Works with any MCP-compatible client
- **No Custom Integrations**: Uses Model Context Protocol standard
- **Real-time Data Access**: AI can query and update data during conversations

For more information on Model Context Protocol, visit [modelcontextprotocol.io](https://modelcontextprotocol.io).

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
