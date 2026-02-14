---
description: 'Instructions for checking for the lastest Go version'
applyTo: '**/*.go,**/go.mod,**/go.sum'
---

## Request

```bash
curl https://go.dev/VERSION?m=text
```

```powershell
Invoke-WebRequest https://go.dev/VERSION?m=text | %{$_.Content}
```

## Response Structure

```text
go1.20.0
time 2026-02-10T01:22:00Z
```
