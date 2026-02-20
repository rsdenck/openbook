$ErrorActionPreference = "Stop"

# --- Configuration ---
$VERSION = "v1.0.0"
$DIST_DIR = "dist"

# --- Clean ---
Write-Host "Cleaning dist directory..." -ForegroundColor Green
if (Test-Path $DIST_DIR) { Remove-Item $DIST_DIR -Recurse -Force }
New-Item -ItemType Directory -Path $DIST_DIR | Out-Null

# --- Targets ---
$targets = @(
    @{OS="linux"; Arch="amd64"}
)

# --- Build Loop ---
foreach ($target in $targets) {
    $os = $target.OS
    $arch = $target.Arch
    $ext = ""
    if ($os -eq "windows") { $ext = ".exe" }
    
    Write-Host "Building for $os/$arch..." -ForegroundColor Cyan
    
    # Set Environment Variables for Go
    $env:GOOS = $os
    $env:GOARCH = $arch
    $env:CGO_ENABLED = "0"
    
    # Build API
    $apiOutput = Join-Path $DIST_DIR "openbook-api-$os-$arch$ext"
    go build -ldflags "-s -w -X main.version=$VERSION" -o $apiOutput ./cmd/api
    
    # Build Worker
    $workerOutput = Join-Path $DIST_DIR "openbook-worker-$os-$arch$ext"
    go build -ldflags "-s -w -X main.version=$VERSION" -o $workerOutput ./cmd/worker
    
    # Package
    $pkgName = "openbook_${VERSION}_${os}_${arch}"
    $pkgDir = Join-Path $DIST_DIR $pkgName
    New-Item -ItemType Directory -Path $pkgDir | Out-Null
    
    Copy-Item $apiOutput -Destination (Join-Path $pkgDir ("openbook-api" + $ext))
    Copy-Item $workerOutput -Destination (Join-Path $pkgDir ("openbook-worker" + $ext))
    Copy-Item "README.md" -Destination $pkgDir
    if (Test-Path "LICENSE") { Copy-Item "LICENSE" -Destination $pkgDir }
    Copy-Item "install.sh" -Destination $pkgDir
    
    # Create Tar
    $tarPath = Join-Path $DIST_DIR "$pkgName.tar.gz"
    # Using tar since it's available (bsdtar)
    # We need to be careful with paths for tar
    $current = Get-Location
    Set-Location $DIST_DIR
    tar -czf "$pkgName.tar.gz" $pkgName
    Set-Location $current
    
    # Cleanup intermediate
    Remove-Item $pkgDir -Recurse -Force
    Remove-Item $apiOutput -Force
    Remove-Item $workerOutput -Force
}

# --- Checksums ---
Write-Host "Generating checksums..." -ForegroundColor Green
$current = Get-Location
Set-Location $DIST_DIR
Get-ChildItem -Filter "*.tar.gz" | ForEach-Object {
    $hash = Get-FileHash $_.Name -Algorithm SHA256
    $hash.Hash.ToLower() + "  " + $_.Name
} | Out-File "checksums.txt" -Encoding ascii
Set-Location $current

Write-Host "Build complete! Artifacts in $DIST_DIR/" -ForegroundColor Green
Get-ChildItem $DIST_DIR
