# Get the latest tag from this repository, the format is v1.0.2
$latestTag = (git describe --tags --abbrev=0).Trim()

# Check if the latest tag matches the expected format
if ($latestTag -match 'v(\d+)\.(\d+)\.(\d+)$') {
    # Extract the major, minor, and patch versions
    $major = [int]$matches[1]
    $minor = [int]$matches[2]
    $patch = [int]$matches[3]

    # Increment the patch version
    $patch++

    # Construct the next version
    $nextVersion = "v$major.$minor.$patch"

    Write-Host "Latest tag: $latestTag, Next version: $nextVersion"
} else {
    Write-Host "Error: Latest tag '$latestTag' does not match the expected format 'v<major>.<minor>.<patch>'"
}

Write-Host "Are you sure you want to trigger the release pipeline with the next version '$nextVersion'? (y/n)"
$confirmation = Read-Host -Prompt "Confirmation"
if ($confirmation -eq 'y') {
    Write-Host "Triggering the release pipeline with the next version '$nextVersion'..."
} else {
    Write-Host "Release pipeline was not triggered."
    exit
}

# set tag
git tag $nextVersion

# push tag
git push origin $nextVersion