$ErrorActionPreference = "Stop"

# Post-tool-use hook - runs after a tool completes
# This is useful for logging, notifications, or cleanup

# Resolve plugin root
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$PluginRoot = (Resolve-Path (Join-Path $ScriptDir "..\..\..")).Path

# Detect OS and select binary
$BinName = if ($IsWindows -or $env:OS -match "Windows") {
    "agentic-ops-windows-amd64.exe"
} elseif ($IsMacOS) {
    if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq "Arm64") {
        "agentic-ops-darwin-arm64"
    } else {
        "agentic-ops-darwin-amd64"
    }
} else {
    "agentic-ops-linux-amd64"
}

$CLI = Join-Path (Join-Path $PluginRoot "bin") $BinName

# Check if CLI exists
if (-not (Test-Path $CLI)) {
    Write-Output '{"permissionDecision":"allow"}'
    exit 0
}

# Read hook input from stdin
try {
    $rawInput = [Console]::In.ReadToEnd()
    if ([string]::IsNullOrWhiteSpace($rawInput)) {
        Write-Output '{"permissionDecision":"allow"}'
        exit 0
    }
    $parsed = $rawInput | ConvertFrom-Json
} catch {
    Write-Output '{"permissionDecision":"allow"}'
    exit 0
}

$toolName = $parsed.toolName
$sessionCwd = $parsed.cwd

# Handle toolArgs
$toolArgs = $parsed.toolArgs
if ($toolArgs -is [string]) {
    try {
        $toolArgs = $toolArgs | ConvertFrom-Json
    } catch {
        $toolArgs = @{}
    }
}

# Convert PSCustomObject to hashtable for proper JSON serialization
function ConvertTo-Hashtable($obj) {
    if ($obj -is [System.Collections.IDictionary]) {
        return $obj
    }
    if ($obj -is [PSCustomObject]) {
        $ht = @{}
        foreach ($prop in $obj.PSObject.Properties) {
            $ht[$prop.Name] = ConvertTo-Hashtable $prop.Value
        }
        return $ht
    }
    return $obj
}

$toolArgsHt = ConvertTo-Hashtable $toolArgs

# Normalize path to relative path from cwd (Copilot passes absolute paths)
if ($toolArgsHt.path -and $sessionCwd) {
    $absPath = $toolArgsHt.path
    if ([System.IO.Path]::IsPathRooted($absPath)) {
        $cwdNorm = $sessionCwd -replace '\\', '/'
        $pathNorm = $absPath -replace '\\', '/'
        if ($pathNorm.StartsWith($cwdNorm, [System.StringComparison]::OrdinalIgnoreCase)) {
            $toolArgsHt.path = $pathNorm.Substring($cwdNorm.Length).TrimStart('/')
        }
    }
}

# Build event JSON
$event = @{
    hook = @{
        type = "postToolUse"
        tool = @{
            name = $toolName
            args = $toolArgsHt
        }
        cwd = $sessionCwd
    }
    tool = @{
        name = $toolName
        args = $toolArgsHt
        hook_type = "postToolUse"
    }
    cwd = $sessionCwd
    timestamp = (Get-Date).ToString("o")
} | ConvertTo-Json -Depth 10 -Compress

# Run agentic-ops CLI
try {
    $result = $event | & $CLI run --event - --dir $sessionCwd 2>&1
    
    try {
        $resultJson = $result | ConvertFrom-Json
        if ($resultJson.permissionDecision) {
            Write-Output ($resultJson | ConvertTo-Json -Compress)
            exit 0
        }
    } catch {}
    
    Write-Output '{"permissionDecision":"allow"}'
} catch {
    Write-Output '{"permissionDecision":"allow"}'
}

exit 0
