param(
  [string]$BinaryName,
  [string]$DistDir
)

$ErrorActionPreference = "Stop"

if (!$BinaryName) { $BinaryName = "ash" }
if (!$DistDir) { $DistDir = "dist" }

# Pick the best exe from dist/
$exe = Join-Path $DistDir "$BinaryName-windows-amd64.exe"
if (!(Test-Path $exe)) {
  $exe = Join-Path $DistDir "$BinaryName-windows-arm64.exe"
}
if (!(Test-Path $exe)) {
  Write-Error "No Windows binary found in $DistDir"
}

$targetDir = Join-Path $env:USERPROFILE "bin"
New-Item -ItemType Directory -Force -Path $targetDir | Out-Null

$dest = Join-Path $targetDir "$BinaryName.exe"
Copy-Item $exe $dest -Force

# Ensure user PATH contains %USERPROFILE%\bin
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $userPath.ToLower().Contains("\bin")) {
  $newPath = "$userPath;$targetDir"
  [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
  Write-Host "Added $targetDir to USER PATH. Open a new terminal to use it."
}

Write-Host "Installed: $dest"
Write-Host "Run: $BinaryName --help"
