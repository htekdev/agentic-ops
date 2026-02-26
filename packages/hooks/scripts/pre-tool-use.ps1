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

# Detect git commit/push commands in shell tools
$commitEvent = $null
$pushEvent = $null

if ($toolName -in @("powershell", "bash", "shell", "cmd")) {
    $command = $toolArgsHt.command
    if (-not $command) { $command = $toolArgsHt.script }
    if (-not $command) { $command = $toolArgsHt.code }
    
    if ($command) {
        # Detect git commit - pattern handles git with flags like -C, --no-pager, etc.
        if ($command -match 'git\b.*\bcommit\b') {
            try {
                Push-Location $sessionCwd
                
                # Get staged files
                $stagedFiles = @()
                $gitStatus = git diff --cached --name-status 2>$null
                if ($gitStatus) {
                    foreach ($line in $gitStatus -split "`n") {
                        if ($line -match '^([AMDRC])\s+(.+)$') {
                            $status = switch ($Matches[1]) {
                                'A' { 'added' }
                                'M' { 'modified' }
                                'D' { 'deleted' }
                                'R' { 'renamed' }
                                'C' { 'copied' }
                                default { 'modified' }
                            }
                            $stagedFiles += @{ path = $Matches[2]; status = $status }
                        }
                    }
                }
                
                # Get commit message from command if present
                $message = ""
                if ($command -match '-m\s+[''"]([^''"]+)[''"]') {
                    $message = $Matches[1]
                } elseif ($command -match '-m\s+(\S+)') {
                    $message = $Matches[1]
                }
                
                # Get current branch
                $branch = git rev-parse --abbrev-ref HEAD 2>$null
                
                $commitEvent = @{
                    sha = "pending"
                    message = $message
                    author = (git config user.email 2>$null)
                    branch = $branch
                    files = $stagedFiles
                }
                
                Pop-Location
            } catch {
                if ((Get-Location).Path -ne $sessionCwd) { Pop-Location }
            }
        }
        
        # Detect git push - pattern handles git with flags like -C, --no-pager, etc.
        if ($command -match 'git\b.*\bpush\b') {
            try {
                Push-Location $sessionCwd
                
                # Get current branch
                $branch = git rev-parse --abbrev-ref HEAD 2>$null
                $ref = "refs/heads/$branch"
                
                # Check if pushing tags
                if ($command -match 'git\b.*\bpush\b.*--tags\b' -or $command -match 'git\b.*\bpush\b.*\borigin\s+(v[\d\.]+|refs/tags/)') {
                    if ($command -match 'git\b.*\bpush\b.*\borigin\s+(v[\d\.]+)') {
                        $ref = "refs/tags/$($Matches[1])"
                    }
                }
                
                # Get current commit
                $currentSha = git rev-parse HEAD 2>$null
                
                $pushEvent = @{
                    ref = $ref
                    before = "0000000000000000000000000000000000000000"
                    after = $currentSha
                }
                
                Pop-Location
            } catch {
                if ((Get-Location).Path -ne $sessionCwd) { Pop-Location }
            }
        }
    }
}

# Build event JSON for the CLI
$event = @{
    hook = @{
        type = "preToolUse"
        tool = @{
            name = $toolName
            args = $toolArgsHt
        }
        cwd = $sessionCwd
    }
    tool = @{
        name = $toolName
        args = $toolArgsHt
        hook_type = "preToolUse"
    }
    cwd = $sessionCwd
    timestamp = (Get-Date).ToString("o")
}

# Add commit event if detected
if ($commitEvent) {
    $event.commit = $commitEvent
}

# Add push event if detected
if ($pushEvent) {
    $event.push = $pushEvent
}

$eventJson = $event | ConvertTo-Json -Depth 10 -Compress

# Run agentic-ops CLI
try {
    $result = $eventJson | & $CLI run --event - --dir $sessionCwd 2>&1
    
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
