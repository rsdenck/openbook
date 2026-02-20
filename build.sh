#!/bin/bash
set -e

# --- Configuration ---
VERSION="v1.0.0"
DIST_DIR="dist"

# --- Colors ---
GREEN='\033[0;32m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }

# --- Clean ---
log_info "Cleaning dist directory..."
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

# --- Targets ---
targets=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
)

# --- Build Loop ---
for target in "${targets[@]}"; do
  os="${target%%/*}"
  arch="${target##*/}"
  ext=""
  if [ "$os" == "windows" ]; then
    ext=".exe"
  fi
  
  log_info "Building for $os/$arch..."
  
  # Build API
  GOOS=$os GOARCH=$arch go build -ldflags "-s -w -X main.version=$VERSION" -o "$DIST_DIR/openbook-api-$os-$arch$ext" ./cmd/api
  
  # Build Worker
  GOOS=$os GOARCH=$arch go build -ldflags "-s -w -X main.version=$VERSION" -o "$DIST_DIR/openbook-worker-$os-$arch$ext" ./cmd/worker
  
  # Package
  pkg_dir="$DIST_DIR/openbook_${VERSION}_${os}_${arch}"
  mkdir -p "$pkg_dir"
  
  cp "$DIST_DIR/openbook-api-$os-$arch$ext" "$pkg_dir/openbook-api$ext"
  cp "$DIST_DIR/openbook-worker-$os-$arch$ext" "$pkg_dir/openbook-worker$ext"
  cp README.md "$pkg_dir/"
  cp LICENSE "$pkg_dir/" || touch "$pkg_dir/LICENSE"
  cp install.sh "$pkg_dir/"
  
  tar -czvf "$DIST_DIR/openbook_${VERSION}_${os}_${arch}.tar.gz" -C "$DIST_DIR" "openbook_${VERSION}_${os}_${arch}"
  
  # Cleanup intermediate files
  rm -rf "$pkg_dir" "$DIST_DIR/openbook-api-$os-$arch$ext" "$DIST_DIR/openbook-worker-$os-$arch$ext"
done

# --- Checksums ---
log_info "Generating checksums..."
cd "$DIST_DIR"
sha256sum *.tar.gz > checksums.txt
cd ..

log_info "Build complete! Artifacts in $DIST_DIR/"
ls -lh "$DIST_DIR"
