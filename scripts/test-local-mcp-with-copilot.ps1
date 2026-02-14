$url = "http://localhost:8080/mcp"

# Test if url returns a valid json response
try {
    $response = Invoke-WebRequest -Uri $url -UseBasicParsing
    $json = $response.Content | ConvertFrom-Json
    Write-Host "MCP server is running and returned a valid JSON response."
} catch {
    Write-Host "Failed to connect to MCP server or invalid JSON response."
    exit 1
}

# Test if command copilot.exe is available
try {
    $copilotPath = (Get-Command copilot.exe).Source
    Write-Host "Copilot executable found at: $copilotPath"
} catch {
    Write-Host "Copilot executable not found. Please ensure it is installed and in the system PATH."
    exit 1
}


$p = @'
Use mcp iot-local-debug to interact with the local MCP server for IoT device management and execute this test case.

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
'@


$mcp_config = '{"mcpServers":{"iot-local-debug":{"type":"http","url":"http://localhost:8080/mcp","headers":{},"tools":["*"]}}}'

# Change to script's directory
Set-Location $PSScriptRoot
Write-Host "Working directory: $(Get-Location)"

Write-Host "Running Copilot with the provided prompt and MCP configuration..."
copilot.exe --interactive $p --additional-mcp-config $mcp_config --model gpt-5-mini --log-level all --allow-all-paths --no-ask-user --no-auto-update --no-custom-instructions