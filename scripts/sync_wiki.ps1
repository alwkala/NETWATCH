# Script: sync_wiki.ps1
# Pushes docs/wiki/ into GitHub Wiki repository (https://github.com/alwkala/NETWATCH.wiki.git)

param(
    [string]$RepoUrl = "https://github.com/alwkala/NETWATCH.wiki.git",
    [string]$WikiSourceDir = "docs/wiki"
)

$ErrorActionPreference = "Stop"

Write-Host "==> Checking GitHub Wiki repository status..." -ForegroundColor Cyan

$tempDir = Join-Path $PSScriptRoot "..\scratch\wiki_sync_temp"
if (Test-Path $tempDir) {
    Remove-Item -Recurse -Force $tempDir
}

try {
    git clone $RepoUrl $tempDir
} catch {
    Write-Host "`n[!] GitHub Wiki repository is not yet initialized on GitHub servers." -ForegroundColor Yellow
    Write-Host "    To initialize it:" -ForegroundColor Yellow
    Write-Host "    1. Open https://github.com/alwkala/NETWATCH/wiki in your browser."
    Write-Host "    2. Click the green 'Create the first page' button and click 'Save page'."
    Write-Host "    3. Re-run this script: .\scripts\sync_wiki.ps1`n"
    exit 1
}

Write-Host "==> Syncing files from $WikiSourceDir to Wiki git repo..." -ForegroundColor Cyan
Copy-Item -Path "$WikiSourceDir\*" -Destination $tempDir -Recurse -Force

Push-Location $tempDir
try {
    git add .
    $status = git status --porcelain
    if ($status) {
        git commit -m "docs(wiki): update NETWATCH documentation suite and navigation chrome"
        git push origin master
        Write-Host "`n[✔] GitHub Wiki successfully updated and live!" -ForegroundColor Green
    } else {
        Write-Host "`n[✔] GitHub Wiki is already up-to-date." -ForegroundColor Green
    }
} finally {
    Pop-Location
    Remove-Item -Recurse -Force $tempDir
}
