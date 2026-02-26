$ErrorActionPreference = "Stop"

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

$CLI = Join-Path $PluginRoot "bin" $BinName

# Check if CLI exists
if (-not (Test-Path $CLI)) {
    # CLI not built yet, allow by default
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
    # Can't parse input, allow by default
    Write-Output '{"permissionDecision":"allow"}'
    exit 0
}

$toolName = $parsed.toolName
$sessionCwd = $parsed.cwd

# Handle toolArgs - may be JSON string or object
$toolArgs = $parsed.toolArgs
if ($toolArgs -is [string]) {
    try {
        $toolArgs = $toolArgs | ConvertFrom-Json
    } catch {
        $toolArgs = @{}
    }
}

# Build event JSON for the CLI
$event = @{
    hook = @{
        type = "preToolUse"
        tool = @{
            name = $toolName
            args = $toolArgs
        }
        cwd = $sessionCwd
    }
    tool = @{
        name = $toolName
        args = $toolArgs
        hook_type = "preToolUse"
    }
    cwd = $sessionCwd
    timestamp = (Get-Date).ToString("o")
} | ConvertTo-Json -Depth 10 -Compress

# Run agentic-ops CLI
try {
    $result = $event | & $CLI run --event - --dir $sessionCwd 2>&1
    
    # Try to parse result as JSON
    try {
        $resultJson = $result | ConvertFrom-Json
        if ($resultJson.permissionDecision) {
            Write-Output ($resultJson | ConvertTo-Json -Compress)
            exit 0
        }
    } catch {
        # Not valid JSON, check for errors
    }
    
    # Default to allow if no valid result
    Write-Output '{"permissionDecision":"allow"}'
} catch {
    # CLI error, allow by default
    Write-Output '{"permissionDecision":"allow"}'
}

exit 0
