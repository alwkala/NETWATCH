# scripts/build_release.ps1
# Automates NETWATCH desktop release packaging, zip creation, and SHA-256 checksum generation.
[CmdletBinding()]
param(
    [string]$Version = "0.2.0-alpha.1",
    [switch]$SkipFrontendBuild
)

$ErrorActionPreference = "Stop"
$RepoRoot = (Get-Item $PSScriptRoot).Parent.FullName
Set-Location $RepoRoot

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host " NETWATCH Release Packager - v$Version" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Build Frontend
if (-not $SkipFrontendBuild) {
    Write-Host "[1/5] Building Frontend (TypeScript and Vite)..." -ForegroundColor Yellow
    Push-Location "$RepoRoot\frontend"
    npm run build
    if ($LASTEXITCODE -ne 0) {
        Pop-Location
        throw "Frontend build failed with exit code $LASTEXITCODE"
    }
    Pop-Location
} else {
    Write-Host "[1/5] Skipping Frontend build (-SkipFrontendBuild specified)..." -ForegroundColor Gray
}

# 2. Prepare Dist Directory
$DistDir = "$RepoRoot\build\dist"
if (Test-Path $DistDir) {
    Remove-Item $DistDir -Recurse -Force
}
New-Item -ItemType Directory -Path $DistDir -Force | Out-Null

# 3. Compile Windows GUI Binary
Write-Host "[2/5] Compiling Windows desktop binary (GUI subsystem suppression)..." -ForegroundColor Yellow
$ExePath = "$DistDir\netwatch.exe"
$ldflags = "-w -s -H windowsgui"
go build -tags desktop,production -ldflags $ldflags -o $ExePath .
if ($LASTEXITCODE -ne 0) {
    throw "Go build failed with exit code $LASTEXITCODE"
}

# 4. Create ZIP Archive
Write-Host "[3/5] Packaging standalone ZIP archive..." -ForegroundColor Yellow
$ZipName = "netwatch-v$Version-windows-amd64.zip"
$ZipPath = "$DistDir\$ZipName"
Compress-Archive -Path $ExePath -DestinationPath $ZipPath -Force

# 5. Compute SHA-256 Hashes
Write-Host "[4/5] Computing SHA-256 cryptographic hashes..." -ForegroundColor Yellow
$ExeHash = (Get-FileHash -Path $ExePath -Algorithm SHA256).Hash
$ZipHash = (Get-FileHash -Path $ZipPath -Algorithm SHA256).Hash

$SumsContent = "$ExeHash *netwatch.exe`r`n$ZipHash *$ZipName`r`n"
$SumsPath = "$DistDir\SHA256SUMS.txt"
Set-Content -Path $SumsPath -Value $SumsContent -Encoding UTF8

Write-Host "[5/5] Release artifacts generated in: $DistDir" -ForegroundColor Green
Write-Host ""
Write-Host "--- GitHub Release Notes Table ---" -ForegroundColor Cyan
Write-Host "### Assets & Integrity (SHA-256)"
Write-Host ""
Write-Host "| Asset | Description | SHA-256 Checksum |"
Write-Host "|---|---|---|"
Write-Host "| $ZipName | Standalone Windows 64-bit Desktop App (.zip) | $ZipHash |"
Write-Host "| netwatch.exe | Direct Windows 64-bit Executable (.exe) | $ExeHash |"
Write-Host "| SHA256SUMS.txt | Authoritative Cryptographic Hashes | (Raw checksum file) |"
Write-Host ""
Write-Host "#### Verification:"
Write-Host '```powershell'
Write-Host "Get-FileHash .\netwatch.exe -Algorithm SHA256"
Write-Host '```'
Write-Host ""
Write-Host "Release packaging completed successfully!" -ForegroundColor Green
