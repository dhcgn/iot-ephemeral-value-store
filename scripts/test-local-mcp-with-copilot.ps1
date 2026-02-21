param(
    [switch]$Verbose,
    [switch]$SkipValidation
)

$ErrorActionPreference = "Stop"
$url = "http://localhost:8080/mcp"

Write-Host "`n=== IoT Ephemeral Value Store - MCP Test Script ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Run npx @modelcontextprotocol/inspector to run a browser based test suite for interactive testing." -ForegroundColor Yellow
Write-Host ""

# Test if MCP server is running and responsive
Write-Host "1. Checking MCP server at $url..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 5
    $json = $response.Content | ConvertFrom-Json
    Write-Host "   ✓ MCP server is running" -ForegroundColor Green
    
    if ($json.server) {
        Write-Host "   Server: $($json.server.name) v$($json.server.version)" -ForegroundColor Gray
    }
    
    if ($json.capabilities.tools.available) {
        Write-Host "   Available tools: $($json.capabilities.tools.available -join ', ')" -ForegroundColor Gray
        
        # Verify required tools exist
        $requiredTools = @("generate_key_pair", "upload_data", "download_data")
        $missingTools = $requiredTools | Where-Object { $_ -notin $json.capabilities.tools.available }
        
        if ($missingTools.Count -gt 0 -and -not $SkipValidation) {
            Write-Host "   ✗ Missing required tools: $($missingTools -join ', ')" -ForegroundColor Red
            exit 1
        }
    }
} catch {
    Write-Host "   ✗ Failed to connect to MCP server: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "   Please ensure the server is running: go run .\main.go" -ForegroundColor Yellow
    exit 1
}

# Test if copilot.exe is available
Write-Host "`n2. Checking for Copilot CLI..." -ForegroundColor Yellow
try {
    $copilotPath = (Get-Command copilot.exe -ErrorAction Stop).Source
    Write-Host "   ✓ Copilot executable found at: $copilotPath" -ForegroundColor Green
    
    # Get version if possible
    try {
        $versionOutput = & copilot.exe --version 2>&1
        if ($versionOutput) {
            Write-Host "   Version: $($versionOutput[0])" -ForegroundColor Gray
        }
    } catch {
        # Version check failed, but copilot exists
    }
} catch {
    Write-Host "   ✗ Copilot executable not found" -ForegroundColor Red
    Write-Host "   Please ensure GitHub Copilot CLI is installed and in your PATH" -ForegroundColor Yellow
    exit 1
}

# Optional: Test REST API endpoints directly
if (-not $SkipValidation) {
    Write-Host "`n3. Validating REST API endpoints..." -ForegroundColor Yellow
    try {
        # Test key generation
        $testKeys = Invoke-RestMethod -Uri "http://localhost:8080/kp" -TimeoutSec 5
        if ($testKeys.'upload-key' -and $testKeys.'download-key') {
            Write-Host "   ✓ Key generation endpoint working" -ForegroundColor Green
            
            # Test upload
            $uploadUrl = "http://localhost:8080/u/$($testKeys.'upload-key')?test=validation"
            $uploadResult = Invoke-RestMethod -Uri $uploadUrl -TimeoutSec 5
            Write-Host "   ✓ Upload endpoint working" -ForegroundColor Green
            
            # Test download
            $downloadUrl = "http://localhost:8080/d/$($testKeys.'download-key')/json"
            $downloadResult = Invoke-RestMethod -Uri $downloadUrl -TimeoutSec 5
            Write-Host "   ✓ Download endpoint working" -ForegroundColor Green
        }
    } catch {
        Write-Host "   ⚠ REST API validation failed: $($_.Exception.Message)" -ForegroundColor Yellow
        Write-Host "   Continuing with MCP test..." -ForegroundColor Gray
    }
}


$p = @'
IMPORTANT: You are ONLY allowed to use MCP tools for testing. Do NOT use direct REST API calls (curl, Invoke-RestMethod, etc.).

Use mcp iot-local-debug to interact with the local MCP server for IoT device management and execute this test plan.

Test Plan: Comprehensive MCP Tool Validation

Step 1: "Generate a key pair for my IoT device"
↓
MCP calls: generate_key_pair (no parameters needed)
↓
Expected: Receive {"upload_key": "...", "download_key": "...", "upload_url": "...", "download_url": "..."}

Step 2: "Upload initial readings"
↓
MCP calls: upload_data with the upload_key from step 1
Parameters: {"tempC": "23.5", "humidity": "45"}
↓
Expected: Data uploaded successfully and parameter_urls returned

Step 3: "Download full JSON"
↓
MCP calls: download_data with the download_key from step 1
↓
Expected: Receive data containing tempC, humidity, and timestamp

Step 4: "Download a single parameter"
↓
MCP calls: download_data with the same download_key
Parameter: "tempC"
↓
Expected: Receive only the tempC value

Step 5: "Patch nested data"
↓
MCP calls: patch_data with the same upload_key
Path: "living_room/sensors"
Parameters: {"co2": "415", "voc": "0.6"}
↓
Expected: Data merged successfully with path reflected in parameter_urls

Step 6: "Verify nested data"
↓
MCP calls: download_data with the same download_key
Parameter: "living_room/sensors/co2"
↓
Expected: Receive value "415"

Step 7: "Patch root-level data"
↓
MCP calls: patch_data with the same upload_key
Path: ""
Parameters: {"status": "ok"}
↓
Expected: Root merged successfully

Step 8: "Verify root-level data"
↓
MCP calls: download_data with the same download_key
Parameter: "status"
↓
Expected: Receive value "ok"

Step 9: "Delete data"
↓
MCP calls: delete_data with the same upload_key
↓
Expected: Success response

Step 10: "Confirm deletion"
↓
MCP calls: download_data with the same download_key
↓
Expected: Error indicating data not found or invalid download key

CONSTRAINT: You MUST use only the MCP tools (generate_key_pair, upload_data, patch_data, download_data, delete_data) from the iot-local-debug server. Do NOT fall back to HTTP requests or REST API calls.

Please execute all steps using MCP tools only and confirm each expected outcome.
'@

$mcp_config = '{"mcpServers":{"iot-local-debug":{"type":"http","url":"http://localhost:8080/mcp","headers":{},"tools":["*"]}}}'

# Change to script's directory
Set-Location $PSScriptRoot

Write-Host "`n=== Running MCP Test with Copilot CLI ===" -ForegroundColor Cyan
Write-Host "Working directory: $(Get-Location)" -ForegroundColor Gray
Write-Host ""

if ($Verbose) {
    Write-Host "MCP Configuration:" -ForegroundColor Yellow
    Write-Host $mcp_config -ForegroundColor Gray
    Write-Host ""
    Write-Host "Test Prompt:" -ForegroundColor Yellow
    Write-Host $p -ForegroundColor Gray
    Write-Host ""
}

Write-Host "Invoking Copilot CLI..." -ForegroundColor Yellow
Write-Host "(This may take a minute as the AI processes the request)" -ForegroundColor Gray
Write-Host ""

$p_quick = @'
IMPORTANT: You are ONLY allowed to use MCP tools for testing. Do NOT use direct REST API calls (curl, Invoke-RestMethod, etc.).

Use mcp iot-local-debug to interact with the local MCP server.

Quick Test: Generate Key Pair Only

Step 1: "Generate a key pair for my IoT device"
↓
MCP calls: generate_key_pair (no parameters needed)
↓
Expected: Receive {"upload_key": "...", "download_key": "..."}

CONSTRAINT: You MUST use only the MCP tool generate_key_pair from the iot-local-debug server.

Please execute the step using MCP tools only and confirm the expected outcome.
'@

$copilotArgs = @(
    "--interactive"
    $p
    "--additional-mcp-config"
    $mcp_config
    "--model"
    "gpt-5-mini"
    "--allow-all-paths"
    "--no-ask-user"
    "--no-auto-update"
    "--no-custom-instructions"
    "--allow-tool"
    "iot-local-debug"
)

if ($Verbose) {
    $copilotArgs += "--log-level"
    $copilotArgs += "all"
}


$startTime = Get-Date

try {
    & copilot.exe @(
        "--prompt"
        $p_quick
        "--additional-mcp-config"
        $mcp_config
        "--model"
        "gpt-5-mini"
        "--allow-all-paths"
        "--no-ask-user"
        "--no-auto-update"
        "--no-custom-instructions"
        # "--allow-tool"
        # "iot-local-debug"
        # "--allow-tool"
        # "iot-local-debug-generate_key_pair"
        # "--allow-tool"
        # "iot-local-debug(*)"
        # "--allow-tool"
        # "generate_key_pair"
        "--yolo" # TODO: Must be changed in future!
    )
    $exitCode = $LASTEXITCODE

    if ($exitCode -ne 0) {
        Write-Host "⚠ Quick test completed with exit code: $exitCode" -ForegroundColor Yellow
        exit $exitCode
    }
} catch {
    Write-Host "`n✗ Error running Copilot CLI (quick test): $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

try {
    & copilot.exe @copilotArgs
    $exitCode = $LASTEXITCODE
    
    $duration = ((Get-Date) - $startTime).TotalSeconds
    
    Write-Host "`n=== Test Completed ===" -ForegroundColor Cyan
    Write-Host "Duration: $([math]::Round($duration, 2)) seconds" -ForegroundColor Gray
    
    if ($exitCode -eq 0) {
        Write-Host "✓ Test completed successfully (exit code: 0)" -ForegroundColor Green
    } else {
        Write-Host "⚠ Test completed with exit code: $exitCode" -ForegroundColor Yellow
    }
    
    exit $exitCode
} catch {
    Write-Host "`n✗ Error running Copilot CLI: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}