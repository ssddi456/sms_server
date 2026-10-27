# SMS Server Remote Deploy Script (SSH + Podman)
# Upload source code to a Linux host and build/deploy there with Podman.
# This flow avoids local docker build and uses mirror base images by default.

param(
    [Parameter(Mandatory = $true)]
    [string]$RemoteHost,
    [string]$RemoteUser = "root",
    [string]$RemotePath = "/opt/sms-server",
    [int]$SshPort = 22,
    [string]$IdentityFile = "",
    [string]$ImageRef = "localhost/sms-server-go:latest",
    [string]$ContainerName = "sms-server",
    [string]$HealthUrl = "http://127.0.0.1:8322/sms/healthz",
    [int]$HealthTimeout = 90,
    [string]$BindAddress = "127.0.0.1",
    [int]$HostPort = 8322,
    [string]$ContainerPort = "8322",
    [string]$GoBaseImage = "docker.m.daocloud.io/library/golang:1.22-alpine",
    [string]$RuntimeBaseImage = "docker.m.daocloud.io/library/alpine:3.20",
    [string]$NodeBaseImage = "docker.m.daocloud.io/library/node:20-alpine"
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

if (-not (Get-Command tar -ErrorAction SilentlyContinue)) {
    Write-Host "✗ tar command not found. Please install tar (Git Bash/bsdtar)." -ForegroundColor Red
    exit 1
}

$ProjectRoot = Split-Path -Parent $PSScriptRoot
$BuildDir = Join-Path $ProjectRoot "build"
$ArchivePath = Join-Path $BuildDir "sms-server-source.tar.gz"
$RemoteTarget = "${RemoteUser}@${RemoteHost}"

if (-not (Test-Path $BuildDir)) {
    New-Item -ItemType Directory -Path $BuildDir -Force | Out-Null
}

$sshArgs = @("-p", "$SshPort")
$scpArgs = @("-P", "$SshPort")
if (-not [string]::IsNullOrWhiteSpace($IdentityFile)) {
    if (-not (Test-Path $IdentityFile)) {
        Write-Host "✗ Identity file not found: $IdentityFile" -ForegroundColor Red
        exit 1
    }
    $sshArgs += @("-i", $IdentityFile)
    $scpArgs += @("-i", $IdentityFile)
}

Write-Host "================================" -ForegroundColor Cyan
Write-Host "SMS Server Remote Deployment (Podman)" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Target: ${RemoteTarget}:$RemotePath" -ForegroundColor Cyan
Write-Host "Image: $ImageRef" -ForegroundColor Cyan
Write-Host "Container: $ContainerName" -ForegroundColor Cyan

Write-Host "`n[1/4] Creating source archive..." -ForegroundColor Yellow
try {
    Push-Location $ProjectRoot
    if (Test-Path $ArchivePath) {
        Remove-Item $ArchivePath -Force
    }
    tar -czf $ArchivePath `
        --exclude ".git" `
        --exclude ".github" `
        --exclude ".vscode" `
        --exclude "build" `
        --exclude "data" `
        --exclude "node_modules" `
        --exclude ".env" `
        --exclude "frontend/static/app" `
        .
    if ($LASTEXITCODE -ne 0) {
        throw "tar failed with exit code $LASTEXITCODE"
    }
}
catch {
    Write-Host "✗ Failed to create source archive: $_" -ForegroundColor Red
    exit 1
}
finally {
    Pop-Location
}

$RemoteArchivePath = "$RemotePath/sms-server-source.tar.gz"

Write-Host "`n[2/4] Uploading source archive..." -ForegroundColor Yellow
try {
    & ssh @sshArgs -o StrictHostKeyChecking=accept-new $RemoteTarget "mkdir -p '$RemotePath'"
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to create remote directory"
    }

    & scp @scpArgs $ArchivePath "${RemoteTarget}:${RemoteArchivePath}"
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to upload source archive"
    }
}
catch {
    Write-Host "✗ Upload failed: $_" -ForegroundColor Red
    exit 1
}

Write-Host "`n[3/4] Building and deploying on remote host..." -ForegroundColor Yellow
try {
    $remoteCommand = @"
set -eu
cd '$RemotePath'
tar -xzf 'sms-server-source.tar.gz'

if [ ! -f '.env' ]; then
  if [ -f '.env.example' ]; then
    cp '.env.example' '.env'
  fi
fi
mkdir -p data

if ! command -v podman >/dev/null 2>&1; then
  echo "podman command not found" >&2
  exit 1
fi
if ! command -v curl >/dev/null 2>&1; then
  echo "curl command not found" >&2
  exit 1
fi

podman build \
  --build-arg GO_BASE_IMAGE='$GoBaseImage' \
  --build-arg RUNTIME_BASE_IMAGE='$RuntimeBaseImage' \
  --build-arg NODE_BASE_IMAGE='$NodeBaseImage' \
  -t '$ImageRef' .

podman rm -f '$ContainerName' >/dev/null 2>&1 || true
podman run -d --name '$ContainerName' \
  --restart unless-stopped \
  --env-file .env \
  -p '${BindAddress}:${HostPort}:${ContainerPort}' \
  -v '$RemotePath/data:/app/data' \
  '$ImageRef'

start_ts=`$(date +%s)
healthy=0
while true; do
  if curl -fsS --max-time 3 '$HealthUrl' >/dev/null 2>&1; then
    healthy=1
    break
  fi
  now_ts=`$(date +%s)
  elapsed=`$((now_ts - start_ts))
  if [ "`$elapsed" -ge "$HealthTimeout" ]; then
    break
  fi
  sleep 2
done

if [ "`$healthy" -eq 1 ]; then
  echo "Deployment successful."
  curl -fsS '$HealthUrl'
  exit 0
fi

echo "Health check failed after ${HealthTimeout}s." >&2
exit 1
"@

    $remoteCommand = $remoteCommand -replace "`r", ""

    & ssh @sshArgs $RemoteTarget $remoteCommand
    if ($LASTEXITCODE -ne 0) {
        throw "Remote deployment failed"
    }
}
catch {
    Write-Host "✗ Remote deployment failed: $_" -ForegroundColor Red
    Write-Host "You can log in and check logs with:" -ForegroundColor Yellow
    Write-Host "podman logs --tail 200 $ContainerName" -ForegroundColor Yellow
    exit 1
}

Write-Host "`n[4/4] Deployment completed." -ForegroundColor Green
Write-Host "================================" -ForegroundColor Cyan
