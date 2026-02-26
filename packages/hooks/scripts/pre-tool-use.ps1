$ErrorActionPreference = "Stop"

# Resolve plugin root
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$PluginRoot = (Resolve-Path (Join-Path $ScriptDir "..\..\..")).Path
$BinDir = Join-Path $PluginRoot "bin"

# Detect OS and architecture
$Arch = if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq "Arm64") {
    "arm64"
} else {
    "amd64"
}

# Select binary based on OS
$BinName = if ($IsWindows -or $env:OS -match "Windows") {
    "agentic-ops-windows-$Arch.exe"
} elseif ($IsMacOS) {
    "agentic-ops-darwin-$Arch"
} else {
    "agentic-ops-linux-$Arch"
}

$CLI = Join-Path $BinDir $BinName
$InstallScript = Join-Path $ScriptDir "install-cli.ps1"

# Check if CLI exists, auto-install if missing
if (-not (Test-Path $CLI)) {
    try {
        if (Test-Path $InstallScript) {
            & $InstallScript -Version "latest" -DestDir $BinDir 2>&1 | Out-Null
        }
    } catch {
        # Install failed, allow by default
        Write-Output '{"permissionDecision":"allow"}'
        exit 0
    }
    
    # Check again after install
    if (-not (Test-Path $CLI)) {
        Write-Output '{"permissionDecision":"allow"}'
        exit 0
    }
} else {
    # CLI exists - check for updates periodically (once per hour)
    $LastCheckFile = Join-Path $BinDir ".last-update-check"
    $ShouldCheck = $true
    
    if (Test-Path $LastCheckFile) {
        $LastCheck = Get-Item $LastCheckFile
        $HoursSinceCheck = ((Get-Date) - $LastCheck.LastWriteTime).TotalHours
        $ShouldCheck = ($HoursSinceCheck -ge 1)
    }
    
    if ($ShouldCheck) {
        try {
            # Update timestamp first to prevent multiple checks
            "" | Out-File -FilePath $LastCheckFile -NoNewline
            
            # Check for updates in background (don't block the hook)
            Start-Job -ScriptBlock {
                param($Script, $Dir)
                & $Script -Version "latest" -DestDir $Dir 2>&1 | Out-Null
            } -ArgumentList $InstallScript, $BinDir | Out-Null
        } catch {
            # Ignore update check errors
        }
    }
}

# Read hook input from stdin
try {
    $rawInput = [Console]::In.ReadToEnd()
    if ([string]::IsNullOrWhiteSpace($rawInput)) {
        Write-Output '{"permissionDecision":"allow"}'
        exit 0
    }
} catch {
    # Can't read input, allow by default
    Write-Output '{"permissionDecision":"allow"}'
    exit 0
}

# Parse input just to get cwd for the CLI
$sessionCwd = ""
try {
    $parsed = $rawInput | ConvertFrom-Json
    $sessionCwd = $parsed.cwd
} catch {
    # Can't parse, try running CLI anyway
}

# Pass raw input directly to CLI with --raw flag
# The CLI will detect event types (git commit, push, file changes, etc.)
try {
    $result = $rawInput | & $CLI run --raw --dir $sessionCwd 2>&1
    
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

