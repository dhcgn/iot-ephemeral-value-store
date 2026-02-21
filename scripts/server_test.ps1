$server = 'http://127.0.0.1:8080' # port from the vscode task

#
# Test Case 1
#

Write-Host 'Test Case 1' -ForegroundColor Cyan

#
# Key generation
#

$kpUrl = $server + '/kp'
$kp = Invoke-RestMethod -Uri $kpUrl

$uploadKey = $kp.'upload-key'
$downloadKey = $kp.'download-key'

if ($uploadKey -eq $null -or $downloadKey -eq $null) {
    Write-Host '🔴 Test failed KeyGen'
    exit
}
else {
    Write-Host '✅ Test passed KeyGen'

}

#
# Upload value
#
$uploadUrl = $server + '/u/' + $uploadKey + '/'
$guid = [guid]::NewGuid().ToString()
$param = "?value=$guid"
$uri = $uploadUrl + $param 
Write-Host $uri -ForegroundColor Yellow
$uploadResult = Invoke-RestMethod -Uri $uri
# $uploadResult

if ($uploadResult.message -eq "Data uploaded successfully") {
    Write-Host '✅ Test passed Upload'
}
else {
    Write-Host '🔴 Test failed Upload'
    exit
}

#
# Download plain value
#
$uri = $uploadResult.parameter_urls[0].value
Write-Host $uri -ForegroundColor Yellow
$downloadResult = Invoke-RestMethod -Uri $uri
# Remove line breaks and whitespaces
$downloadResult = $downloadResult -replace '\r\n', ''
$downloadResult = $downloadResult -replace '\s', ''

if ($downloadResult -eq $guid) {
    Write-Host '✅ Test passed PLAIN'
}
else {
    Write-Host '🔴 Test failed PLAIN'
    Write-Host "Expected: >$guid<"
    Write-Host "Actual: >$downloadResult<"
}

#
# Download json value
#
$uri = $uploadResult.download_url
Write-Host $uri -ForegroundColor Yellow
$downloadResult = Invoke-RestMethod -Uri $uri

if ($downloadResult.value -eq $guid) {
    Write-Host '✅ Test passed JSON'
}
else {
    Write-Host '🔴 Test failed JSON'
    Write-Host "Expected: >$guid<"
    Write-Host "Actual: >$($downloadResult.value)<"
}

#
# Test Case 2
#

Write-Host 'Test Case 2' -ForegroundColor Cyan

#
# Key generation
#

$kpUrl = $server + '/kp'
$kp = Invoke-RestMethod -Uri $kpUrl

$uploadKey = $kp.'upload-key'
$downloadKey = $kp.'download-key'

if ($uploadKey -eq $null -or $downloadKey -eq $null) {
    Write-Host '🔴 Test failed KeyGen'
    exit
}
else {
    Write-Host '✅ Test passed KeyGen'

}

#
# Upload patch value #1
#

$uploadUrl = $server + '/patch/' + $uploadKey + '/1/2'
$guid1 = [guid]::NewGuid().ToString()
$param = "?value=$guid1"
$uri = $uploadUrl + $param 
Write-Host $uri -ForegroundColor Yellow
$uploadResult = Invoke-RestMethod -Uri $uri

if ($uploadResult.message -eq "Data uploaded successfully") {
    Write-Host '✅ Test passed Upload NESTED 1'
}
else {
    Write-Host '🔴 Test failed Upload NESTED 1'
    exit
}

#
# Upload patch value add #1
#

$uploadUrl = $server + '/patch/' + $uploadKey + '/1/2'
$guid1a = [guid]::NewGuid().ToString()
$param = "?valueA=$guid1a"
$uri = $uploadUrl + $param 
Write-Host $uri -ForegroundColor Yellow
$uploadResult = Invoke-RestMethod -Uri $uri

if ($uploadResult.message -eq "Data uploaded successfully") {
    Write-Host '✅ Test passed Upload NESTED 1a'
}
else {
    Write-Host '🔴 Test failed Upload NESTED 1a'
    exit
}

#
# Upload patch value #2
#

$uploadUrl = $server + '/patch/' + $uploadKey + '/2/3'
$guid2 = [guid]::NewGuid().ToString()
$param = "?value=$guid2"
$uri = $uploadUrl + $param 
Write-Host $uri -ForegroundColor Yellow
$uploadResult = Invoke-RestMethod -Uri $uri

if ($uploadResult.message -eq "Data uploaded successfully") {
    Write-Host '✅ Test passed Upload NESTED 2'
}
else {
    Write-Host '🔴 Test failed Upload NESTED 2'
    exit
}

#
# Download json value
#
$uri = $uploadResult.download_url
Write-Host $uri -ForegroundColor Yellow
$downloadResult = Invoke-RestMethod -Uri $uri

<#
Result:
{
  "1": {
    "2": {
      "timestamp": "2024-06-14T19:48:45Z",
      "value": "ca9f804b-9dbf-4ad6-a0ac-8e3dff8e82e1"
    }
  },
  "2": {
    "3": {
      "timestamp": "2024-06-14T19:48:45Z",
      "value": "9a0ca7c6-543a-407d-99e9-a09a05216819"
    }
  }
}
#>

$actual = $downloadResult.'1'.'2'.value
if ($actual -eq $guid1) {
    Write-Host '✅ Test passed JSON NESTED 1'
}
else {
    Write-Host '🔴 Test failed JSON NESTED 1'
    Write-Host "Expected: >$guid1<"
    Write-Host "Actual: >$actual<"
}

$actual = $downloadResult.'1'.'2'.valueA
if ($actual -eq $guid1a) {
    Write-Host '✅ Test passed JSON NESTED 1a'
}
else {
    Write-Host '🔴 Test failed JSON NESTED 1a'
    Write-Host "Expected: >$guid1<"
    Write-Host "Actual: >$actual<"
}

$actual = $downloadResult.'2'.'3'.value
if ($actual -eq $guid2) {
    Write-Host '✅ Test passed JSON NESTED 2'
}
else {
    Write-Host '🔴 Test failed JSON NESTED 2'
    Write-Host "Expected: >$guid2<"
    Write-Host "Actual: >$actual<"
}