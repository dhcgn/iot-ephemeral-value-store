# Test Scripts for IoT Ephemeral Value Store

## test-local-mcp-with-copilot.ps1

A comprehensive test script for validating the MCP (Model Context Protocol) integration with GitHub Copilot CLI.

### Features

✅ **Pre-flight Checks**
- Validates MCP server is running and responsive
- Verifies required MCP tools are available
- Checks GitHub Copilot CLI is installed
- Optional REST API endpoint validation

✅ **Improved Error Handling**
- Clear error messages with color coding
- Timeout protection (5 seconds per check)
- Graceful failure with helpful suggestions

✅ **Better Output**
- Color-coded status messages (Green ✓, Red ✗, Yellow ⚠)
- Duration tracking
- Verbose mode for debugging

### Usage

#### Basic Usage
```powershell
.\scripts\test-local-mcp-with-copilot.ps1
```

#### With Verbose Output
```powershell
.\scripts\test-local-mcp-with-copilot.ps1 -Verbose
```

#### Skip Pre-flight Validation
```powershell
.\scripts\test-local-mcp-with-copilot.ps1 -SkipValidation
```

### Prerequisites

1. **IoT Ephemeral Value Store Server** must be running:
   ```powershell
   cd C:\dev\iot-ephemeral-value-store
   go run .\main.go
   ```
   Server should be accessible at `http://localhost:8080`

2. **GitHub Copilot CLI** must be installed and in PATH:
   ```bash
   copilot --version
   ```

### What It Tests

The script executes a complete IoT data workflow through MCP:

1. **Generate Key Pair**
   - Calls `generate_key_pair` tool (no parameters)
   - Receives upload and download keys

2. **Upload Sensor Data**
   - Calls `upload_data` tool with upload key
   - Stores temperature (23.5°C) and humidity (45%)

3. **Download Sensor Data**
   - Calls `download_data` tool with download key
   - Retrieves and verifies the stored data

### Exit Codes

- `0` - Test completed successfully
- `1` - Server not running, Copilot not found, or test failed

### Troubleshooting

#### "Failed to connect to MCP server"
- Ensure the server is running: `go run .\main.go`
- Check the server is listening on port 8080
- Verify firewall settings

#### "Copilot executable not found"
- Install GitHub Copilot CLI
- Ensure `copilot.exe` is in your system PATH

#### "Missing required tools"
- Restart the server to reload tool registrations
- Use `-SkipValidation` to bypass this check (not recommended)

#### "object schema missing properties" error
- This was fixed by changing `GenerateKeyPairInput struct{}` to `params any`
- Ensure you're running the latest code version
- Rebuild and restart the server

### Examples

#### Successful Run
```
=== IoT Ephemeral Value Store - MCP Test Script ===

1. Checking MCP server at http://localhost:8080/mcp...
   ✓ MCP server is running
   Server: iot-ephemeral-value-store vdev
   Available tools: generate_key_pair, upload_data, patch_data, download_data, delete_data

2. Checking for Copilot CLI...
   ✓ Copilot executable found at: C:\Program Files\GitHub Copilot CLI\copilot.exe

3. Validating REST API endpoints...
   ✓ Key generation endpoint working
   ✓ Upload endpoint working
   ✓ Download endpoint working

=== Running MCP Test with Copilot CLI ===
...

=== Test Completed ===
Duration: 12.45 seconds
✓ Test completed successfully (exit code: 0)
```

### Development

To modify the test case, edit the `$p` variable in the script. The current test validates:
- Key generation (parameter-less tool)
- Data upload with parameters
- Data download and retrieval

### Related Documentation

- [MCP Server Documentation](../README.md)
- [IoT Ephemeral Value Store README](../README.md)
- [Technical Details](../README.TechDetails.md)
