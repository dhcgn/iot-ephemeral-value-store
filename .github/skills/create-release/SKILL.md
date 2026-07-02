---
name: create-release
description: 'Create a new semantic version release for the iot-ephemeral-value-store repository. Use when user asks to create a release, bump the version, tag a new version, or run the release script.'
license: MIT
allowed-tools: Bash
---

# Create Release

## Overview

This repository uses semantic versioning with git tags. Releases are created by:
1. Incrementing the patch version of the latest tag
2. Tagging the latest commit on `main`
3. Pushing the tag to GitHub (which triggers CI to build and publish)

## Script

The repository includes a PowerShell script that automates version bumping:

```
scripts/increase_sem_version_git_tag.ps1
```

This script:
- Reads the latest tag (e.g. `v1.0.21`)
- Increments the patch version (`v1.0.22`)
- Creates the new local git tag

## Full Workflow

### Step 1: Ensure main is up to date

```bash
git fetch origin main:refs/remotes/origin/main
git log --oneline origin/main | head -5
```

### Step 2: Run the version bump script

```bash
pwsh scripts/increase_sem_version_git_tag.ps1
```

The script prints the new version (e.g. `New version tag: v1.0.22`) and creates the tag **on the current HEAD**. If the tag should point to `origin/main`'s latest commit instead, delete and recreate it:

```bash
# Get main's latest SHA
MAIN_SHA=$(git rev-parse origin/main)

# Delete the tag just created (if it pointed to wrong commit)
git tag -d v1.0.22

# Recreate on the correct commit
git tag v1.0.22 $MAIN_SHA
```

### Step 3: Push the tag

```bash
git push origin v1.0.22
```

> **Note:** The sandbox agent cannot run `git push` directly. Use `engine-tools-report_progress` to push branch changes, or instruct the user to run the push command locally.

### Step 4: Create a GitHub Release (optional)

After the tag is pushed, create a GitHub release:

```bash
# Without auto-generated notes (avoids GraphQL permission issues)
gh release create v1.0.22 --title "v1.0.22" --notes "Release v1.0.22"

# With auto-generated notes (requires GraphQL access)
gh release create v1.0.22 --title "v1.0.22" --generate-notes
```

## Version Format

Tags follow the format `vMAJOR.MINOR.PATCH` (e.g. `v1.0.22`).

The script only increments the **patch** version. To bump major or minor versions, edit the tag manually:

```bash
# Example: bump minor version
git tag v1.1.0 $(git rev-parse origin/main)
git push origin v1.1.0
```

## Checking the Latest Tag

```bash
git describe --tags --abbrev=0
```

## Common Issues

| Problem | Solution |
|---|---|
| `HTTP 403` on `gh release create --generate-notes` | Use `--notes "..."` instead (avoids GraphQL) |
| `git push` returns 403 | Push from local machine or use a token with write access |
| Tag points to wrong commit | `git tag -d <tag>` then recreate on correct SHA |
| PowerShell not available | Install with `sudo apt install powershell` or use `pwsh` |
