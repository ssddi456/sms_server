# SMS Server Docker Deploy Script
# This script runs tests, builds Docker image, packages deployment artifacts,
# and optionally uploads the package with HMAC-SHA256 authentication.

param(
    [string]$ImageName = "sms-server-go",
    [string]$ImageTag = "latest",
    [string]$UploadUrl = "",
    [string]$SecretKey = "",
    [switch]$SkipTests,
    [switch]$SkipUpload
)

$ErrorActionPreference = "Stop"

$ProjectRoot = Split-Path -Parent $PSScriptRoot
$BuildDir = Join-Path $ProjectRoot "build"
$PackageDir = Join-Path $BuildDir "deploy-package"
$ArchivePath = Join-Path $BuildDir "sms-server-deploy.tar.gz"
$ImageTarPath = Join-Path $PackageDir "sms-server-image.tar"
$ImageRef = "${ImageName}:${ImageTag}"

Write-Host "================================" -ForegroundColor Cyan
Write-Host "SMS Server Docker Deployment" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Image: $ImageRef" -ForegroundColor Cyan

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "✗ docker command not found. Please install Docker first." -ForegroundColor Red
    exit 1
}

if (-not (Get-Command tar -ErrorAction SilentlyContinue)) {
    Write-Host "✗ tar command not found. Please install Git Bash, WSL, or GNU tar." -ForegroundColor Red
    exit 1
}

try {
    if (Test-Path $BuildDir) {
        Remove-Item $BuildDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $PackageDir -Force | Out-Null
}
catch {
    Write-Host "✗ Failed to initialize build directory: $_" -ForegroundColor Red
    exit 1
}

# Step 1: Run tests
if (-not $SkipTests) {
    Write-Host "`n[1/5] Running go test ./..." -ForegroundColor Yellow
    try {
        Push-Location $ProjectRoot
        go test ./...
        if ($LASTEXITCODE -ne 0) {
            throw "go test failed with exit code $LASTEXITCODE"
        }
        Write-Host "✓ Tests passed" -ForegroundColor Green
    }
    catch {
        Write-Host "✗ Tests failed: $_" -ForegroundColor Red
        exit 1
    }
    finally {
        Pop-Location
    }
}
else {
    Write-Host "`n[1/5] Skip tests (--SkipTests)" -ForegroundColor Yellow
}

# Step 2: Build docker image
Write-Host "`n[2/5] Building docker image..." -ForegroundColor Yellow
try {
    Push-Location $ProjectRoot
    docker build -t $ImageRef .
    if ($LASTEXITCODE -ne 0) {
        throw "docker build failed with exit code $LASTEXITCODE"
    }
    Write-Host "✓ Docker image built: $ImageRef" -ForegroundColor Green
}
catch {
    Write-Host "✗ Docker build failed: $_" -ForegroundColor Red
    exit 1
}
finally {
    Pop-Location
}

# Step 3: Export image and package deployment files
Write-Host "`n[3/5] Exporting image and preparing deployment package..." -ForegroundColor Yellow
try {
    Push-Location $ProjectRoot

    docker save -o $ImageTarPath $ImageRef
    if ($LASTEXITCODE -ne 0) {
        throw "docker save failed with exit code $LASTEXITCODE"
    }

    $PackageDeployDir = Join-Path $PackageDir "deploy"
    $PackageDockerDir = Join-Path $PackageDeployDir "docker"
    $PackageNginxDir = Join-Path $PackageDeployDir "nginx"
    New-Item -ItemType Directory -Path $PackageDockerDir -Force | Out-Null
    New-Item -ItemType Directory -Path $PackageNginxDir -Force | Out-Null

    Copy-Item (Join-Path $ProjectRoot "deploy\docker\docker-compose.prod.yml") (Join-Path $PackageDockerDir "docker-compose.prod.yml")
    Copy-Item (Join-Path $ProjectRoot "deploy\docker\deploy.sh") (Join-Path $PackageDockerDir "deploy.sh")
    Copy-Item (Join-Path $ProjectRoot "deploy\docker\rollback.sh") (Join-Path $PackageDockerDir "rollback.sh")
    Copy-Item (Join-Path $ProjectRoot ".env.example") (Join-Path $PackageDir ".env.example")
    Copy-Item (Join-Path $ProjectRoot "deploy\nginx\sms_subpath.conf") (Join-Path $PackageNginxDir "sms_subpath.conf")

    $manifestPath = Join-Path $PackageDir "manifest.txt"
    @(
        "project=sms_server",
        "image=$ImageRef",
        "created_at=$([DateTime]::UtcNow.ToString("o"))",
        "notes=Run 'docker load -i sms-server-image.tar' then deploy with deploy/docker/docker-compose.prod.yml"
    ) | Set-Content -Path $manifestPath -Encoding UTF8

    Write-Host "✓ Deployment package prepared" -ForegroundColor Green
}
catch {
    Write-Host "✗ Package preparation failed: $_" -ForegroundColor Red
    exit 1
}
finally {
    Pop-Location
}

# Step 4: Create tar.gz package
Write-Host "`n[4/5] Creating deployment archive..." -ForegroundColor Yellow
try {
    Push-Location $BuildDir
    tar -czf "sms-server-deploy.tar.gz" -C "deploy-package" .
    if ($LASTEXITCODE -ne 0) {
        throw "tar failed with exit code $LASTEXITCODE"
    }
    Write-Host "✓ Archive created: $ArchivePath" -ForegroundColor Green
}
catch {
    Write-Host "✗ Archive creation failed: $_" -ForegroundColor Red
    exit 1
}
finally {
    Pop-Location
}

# Step 5: Optional upload with HMAC signature
$shouldUpload = (-not $SkipUpload) -and (-not [string]::IsNullOrWhiteSpace($UploadUrl)) -and (-not [string]::IsNullOrWhiteSpace($SecretKey))
if ($shouldUpload) {
    Write-Host "`n[5/5] Uploading deployment archive..." -ForegroundColor Yellow
    Write-Host "URL: $UploadUrl" -ForegroundColor Cyan

    try {
        $FileHash = Get-FileHash -Path $ArchivePath -Algorithm SHA256
        $Checksum = $FileHash.Hash.ToLower()

        $HmacSha256 = New-Object System.Security.Cryptography.HMACSHA256
        $HmacSha256.Key = [Text.Encoding]::UTF8.GetBytes($SecretKey)
        $SignatureBytes = $HmacSha256.ComputeHash([Text.Encoding]::UTF8.GetBytes($Checksum))
        $Signature = [Convert]::ToBase64String($SignatureBytes)

        $Boundary = "----" + [System.Guid]::NewGuid().ToString().Replace("-", "")
        $ContentType = "multipart/form-data; boundary=$Boundary"

        $FileBytes = [System.IO.File]::ReadAllBytes($ArchivePath)
        $FileName = [System.IO.Path]::GetFileName($ArchivePath)

        $MemoryStream = New-Object System.IO.MemoryStream
        $Writer = New-Object System.IO.StreamWriter($MemoryStream)

        $Writer.Write("--$Boundary`r`n")
        $Writer.Write("Content-Disposition: form-data; name=""checksum""`r`n`r`n")
        $Writer.Write("$Checksum`r`n")

        $Writer.Write("--$Boundary`r`n")
        $Writer.Write("Content-Disposition: form-data; name=""signature""`r`n`r`n")
        $Writer.Write("$Signature`r`n")

        $Writer.Write("--$Boundary`r`n")
        $Writer.Write("Content-Disposition: form-data; name=""image""`r`n`r`n")
        $Writer.Write("$ImageRef`r`n")

        $Writer.Write("--$Boundary`r`n")
        $Writer.Write("Content-Disposition: form-data; name=""file""; filename=""$FileName""`r`n")
        $Writer.Write("Content-Type: application/gzip`r`n`r`n")
        $Writer.Flush()

        $MemoryStream.Write($FileBytes, 0, $FileBytes.Length)

        $Writer.Write("`r`n--$Boundary--`r`n")
        $Writer.Flush()

        $Body = $MemoryStream.ToArray()
        $Writer.Dispose()
        $MemoryStream.Dispose()

        $UploadResponse = Invoke-WebRequest -Uri $UploadUrl `
            -Method Post `
            -ContentType $ContentType `
            -Body $Body `
            -UseBasicParsing `
            -ErrorAction Stop

        if ($UploadResponse.StatusCode -eq 200 -or $UploadResponse.StatusCode -eq 201) {
            Write-Host "✓ Upload successful (HTTP $($UploadResponse.StatusCode))" -ForegroundColor Green
            Write-Host "Response: $($UploadResponse.Content)" -ForegroundColor Cyan
        }
        else {
            Write-Host "✗ Upload failed with status code $($UploadResponse.StatusCode)" -ForegroundColor Red
            Write-Host "Response: $($UploadResponse.Content)" -ForegroundColor Yellow
            exit 1
        }
    }
    catch {
        Write-Host "✗ Upload error: $($_.Exception.Message)" -ForegroundColor Red
        exit 1
    }
}
else {
    Write-Host "`n[5/5] Skip upload (missing UploadUrl/SecretKey or --SkipUpload used)" -ForegroundColor Yellow
}

Write-Host "`n================================" -ForegroundColor Cyan
Write-Host "Deployment package ready" -ForegroundColor Green
Write-Host "Archive: $ArchivePath" -ForegroundColor Cyan
Write-Host "Image: $ImageRef" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
