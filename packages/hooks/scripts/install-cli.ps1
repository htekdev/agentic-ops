#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Downloads the agentic-ops CLI binary from GitHub releases.
    
.DESCRIPTION
    This script downloads the appropriate agentic-ops-cli binary for the current
    platform from the GitHub releases page. It's automatically run by the plugin
    hooks when the CLI is not found.
    
.PARAMETER Version
    The version to download. Defaults to "latest".
    
.PARAMETER DestDir
    The destination directory for the binary. Defaults to the plugin's bin/ folder.
#>

param(
    [string]$Version = "latest",
    [string]$DestDir = ""
)

$ErrorActionPreference = "Stop"

# Determine plugin root
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$PluginRoot = (Resolve-Path (Join-Path (Join-Path (Join-Path $ScriptDir "..") "..") "..")).Path

if (-not $DestDir) {
    $DestDir = Join-Path $PluginRoot "bin"
}

# Create bin directory if it doesn't exist
if (-not (Test-Path $DestDir)) {
    New-Item -ItemType Directory -Path $DestDir -Force | Out-Null
}

# Detect platform
$OS = if ($IsWindows -or $env:OS -match "Windows") {
    "windows"
} elseif ($IsMacOS) {
    "darwin"
} else {
    "linux"
}

$Arch = if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq "Arm64") {
    "arm64"
} else {
    "amd64"
}

$Ext = if ($OS -eq "windows") { ".exe" } else { "" }

$BinaryName = "agentic-ops-$OS-$Arch$Ext"

# Determine download URL
$RepoOwner = "htekdev"
$RepoName = "agentic-ops-cli"

if ($Version -eq "latest") {
    $ApiUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
    try {
        $Release = Invoke-RestMethod -Uri $ApiUrl -Headers @{ "Accept" = "application/vnd.github.v3+json" }
        $Version = $Release.tag_name
    } catch {
        Write-Error "Failed to fetch latest release: $_"
        exit 1
    }
}

$DownloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$Version/$BinaryName"

$DestPath = Join-Path $DestDir $BinaryName

Write-Host "Downloading agentic-ops CLI $Version for $OS/$Arch..."
Write-Host "URL: $DownloadUrl"

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $DestPath -UseBasicParsing
    Write-Host "Downloaded to: $DestPath"
    
    # Make executable on Unix
    if ($OS -ne "windows") {
        chmod +x $DestPath
    }
    
    # Verify binary works
    $TestOutput = & $DestPath version 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Verified CLI: $TestOutput"
    } else {
        Write-Warning "CLI installed but version check returned exit code $LASTEXITCODE"
    }
    
    Write-Host "agentic-ops CLI installed successfully!"
} catch {
    Write-Error "Failed to download CLI: $_"
    exit 1
}
