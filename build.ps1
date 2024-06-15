$restoreGOOS = $env:GOOS

# Set up the environment for cross-compilation
$env:GOOS = "linux"
$env:GOARCH = "amd64"

# Define the output directory and binary name
$outputDir = ".\build"
$binaryName = "iot-ephemeral-value-store-server"

# Check if the output directory exists, if not, create it
if (-Not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir
}

# Build the application
Write-Host "Building application for Linux AMD64..."

$env:CGO_ENABLED = 0
go build -ldflags="-w -s" -o "$outputDir/$binaryName" .

# Check if the build was successful
if ($LASTEXITCODE -eq 0) {
    Write-Host "Build completed successfully. Binary located at $outputDir/$binaryName"
} else {
    Write-Host "Build failed with exit code $LASTEXITCODE"
    exit $LASTEXITCODE
}

# Restore the original GOOS value
$env:GOOS = $restoreGOOS