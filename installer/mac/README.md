```bash
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64  go build -o oadin -ldflags="-s -w"  cmd/cli/main.go

mkdir -p pkgroot/OADIN
mv oadin pkgroot/OADIN/oadin
mkdir -p scripts
# Copy the postinstall script to the scripts directory
chmod +x scripts/postinstall
pkgbuild --identifier com.intel.oadin --version "0.4.0" --install-location /Users/Shared/OADIN --root pkgroot/OADIN --scripts ./scripts oadin.pkg

```