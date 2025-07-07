```bash
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64  go build -o aog -ldflags="-s -w"  cmd/cli/main.go

mkdir -p pkgroot/AOG
mv aog pkgroot/AOG/aog
mkdir -p scripts
# Copy the postinstall script to the scripts directory
chmod +x scripts/postinstall
pkgbuild --identifier com.intel.aog --version "0.4.0" --install-location /Users/Shared/AOG --root pkgroot/AOG --scripts ./scripts aog.pkg

```