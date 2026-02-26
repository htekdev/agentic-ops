#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Downloads the agentic-ops CLI binary from GitHub releases.
    
.DESCRIPTION
    This script downloads the appropriate agentic-ops-cli binary for the current
    platform from the GitHub releases page. It's automatically run by the plugin
    hooks when the CLI is not found or needs updating.
    
.PARAMETER Version
    The version to download. Defaults to "latest".
    
.PARAMETER DestDir
    The destination directory for the binary. Defaults to the plugin's bin/ folder.
    
.PARAMETER CheckOnly
    If set, only checks if update is available without downloading.
#>

param(
    [string]$Version = "latest",
    [string]$DestDir = "",
    [switch]$CheckOnly
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
$VersionFile = Join-Path $DestDir ".version"

# Determine download URL and get latest version
$RepoOwner = "htekdev"
$RepoName = "agentic-ops-cli"

$LatestVersion = $null
if ($Version -eq "latest") {
    $ApiUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
    try {
        $ProgressPreference = 'SilentlyContinue'
        $Release = Invoke-RestMethod -Uri $ApiUrl -Headers @{ "Accept" = "application/vnd.github.v3+json" } -TimeoutSec 5
        $LatestVersion = $Release.tag_name
        $Version = $LatestVersion
    } catch {
        if ($CheckOnly) {
            # Can't check, assume no update
            Write-Output '{"updateAvailable":false,"error":"Failed to check for updates"}'
            exit 0
        }
        Write-Error "Failed to fetch latest release: $_"
        exit 1
    }
}

# Check if update is needed
$CurrentVersion = $null
if (Test-Path $VersionFile) {
    $CurrentVersion = Get-Content $VersionFile -Raw
    $CurrentVersion = $CurrentVersion.Trim()
}

if ($CheckOnly) {
    $UpdateAvailable = ($CurrentVersion -ne $LatestVersion) -and ($null -ne $LatestVersion)
    $result = @{
        updateAvailable = $UpdateAvailable
        currentVersion = $CurrentVersion
        latestVersion = $LatestVersion
    } | ConvertTo-Json -Compress
    Write-Output $result
    exit 0
}

# Skip if already up to date
$DestPath = Join-Path $DestDir $BinaryName
if ((Test-Path $DestPath) -and ($CurrentVersion -eq $Version)) {
    Write-Host "agentic-ops CLI $Version is already installed"
    exit 0
}

$DownloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$Version/$BinaryName"

Write-Host "Downloading agentic-ops CLI $Version for $OS/$Arch..."
Write-Host "URL: $DownloadUrl"

try {
    $ProgressPreference = 'SilentlyContinue'
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $DestPath -UseBasicParsing -TimeoutSec 60
    Write-Host "Downloaded to: $DestPath"
    
    # Make executable on Unix
    if ($OS -ne "windows") {
        chmod +x $DestPath
    }
    
    # Save version info
    $Version | Out-File -FilePath $VersionFile -NoNewline -Encoding utf8
    
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
