# SMS Server Remote Deploy Script (Docker)
# Build deployment package locally, upload to Linux host, and execute remote deployment.

param(
    [Parameter(Mandatory = $true)]
    [string]$RemoteHost,
    [Parameter(Mandatory = $true)]
    [string]$RemoteUser,
    [string]$RemotePath = "/opt/sms-server",
    [int]$SshPort = 22,
    [string]$IdentityFile = "",
    [string]$ImageName = "sms-server-go",
    [string]$ImageTag = "latest",
    [string]$HealthUrl = "http://127.0.0.1:8322/sms/healthz",
    [int]$HealthTimeout = 60,
    [switch]$SkipTests
)

$ErrorActionPreference = "Stop"

if (-not (Get-Command ssh -ErrorAction SilentlyContinue)) {
    Write-Host "✗ ssh command not found. Please install OpenSSH client." -ForegroundColor Red
    exit 1
}

if (-not (Get-Command scp -ErrorAction SilentlyContinue)) {
    Write-Host "✗ scp command not found. Please install OpenSSH client." -ForegroundColor Red
    exit 1
}

$ProjectRoot = Split-Path -Parent $PSScriptRoot
$DeployScriptPath = Join-Path $PSScriptRoot "deploy.ps1"
$ArchivePath = Join-Path $ProjectRoot "build\sms-server-deploy.tar.gz"
$ImageRef = "${ImageName}:${ImageTag}"
$RemoteTarget = "${RemoteUser}@${RemoteHost}"

Write-Host "================================" -ForegroundColor Cyan
Write-Host "SMS Server Remote Deployment" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Target: ${RemoteTarget}:$RemotePath" -ForegroundColor Cyan
Write-Host "Image: $ImageRef" -ForegroundColor Cyan

Write-Host "`n[1/3] Building local deployment package..." -ForegroundColor Yellow
try {
    $deployArgs = @(
        "-ExecutionPolicy", "Bypass",
        "-File", $DeployScriptPath,
        "-ImageName", $ImageName,
        "-ImageTag", $ImageTag,
        "-SkipUpload"
    )
    if ($SkipTests) {
        $deployArgs += "-SkipTests"
    }

    & powershell @deployArgs
    if ($LASTEXITCODE -ne 0) {
        throw "Local deploy packaging failed with exit code $LASTEXITCODE"
    }
}
catch {
    Write-Host "✗ Local package build failed: $_" -ForegroundColor Red
    exit 1
}

if (-not (Test-Path $ArchivePath)) {
    Write-Host "✗ Deployment archive not found: $ArchivePath" -ForegroundColor Red
    exit 1
}

$sshArgs = @("-p", "$SshPort")
$scpArgs = @("-P", "$SshPort")
if (-not [string]::IsNullOrWhiteSpace($IdentityFile)) {
    $sshArgs += @("-i", $IdentityFile)
    $scpArgs += @("-i", $IdentityFile)
}

$RemoteArchivePath = "$RemotePath/sms-server-deploy.tar.gz"

Write-Host "`n[2/3] Uploading deployment archive..." -ForegroundColor Yellow
try {
    & ssh @sshArgs $RemoteTarget "mkdir -p '$RemotePath'"
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to create remote directory"
    }

    & scp @scpArgs $ArchivePath "${RemoteTarget}:${RemoteArchivePath}"
    if ($LASTEXITCODE -ne 0) {
        throw "scp upload failed"
    }
}
catch {
    Write-Host "✗ Upload failed: $_" -ForegroundColor Red
    exit 1
}

Write-Host "`n[3/3] Running remote deployment..." -ForegroundColor Yellow
try {
    $remoteCommand = @"
set -euo pipefail
cd '$RemotePath'
tar -xzf sms-server-deploy.tar.gz
chmod +x deploy/docker/deploy.sh deploy/docker/rollback.sh
./deploy/docker/deploy.sh --image '$ImageRef' --health-url '$HealthUrl' --health-timeout $HealthTimeout
"@

    & ssh @sshArgs $RemoteTarget $remoteCommand
    if ($LASTEXITCODE -ne 0) {
        throw "Remote deployment failed"
    }
}
catch {
    Write-Host "✗ Remote deployment failed: $_" -ForegroundColor Red
    Write-Host "You can log in and run rollback manually:" -ForegroundColor Yellow
    Write-Host "cd $RemotePath && ./deploy/docker/rollback.sh" -ForegroundColor Yellow
    exit 1
}

Write-Host "`n================================" -ForegroundColor Cyan
Write-Host "Remote deployment completed successfully." -ForegroundColor Green
Write-Host "================================" -ForegroundColor Cyan
