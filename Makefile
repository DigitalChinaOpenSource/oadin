GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

build-all:
ifeq ($(GOOS),windows)
	$(MAKE) build-cli-win
else ifeq ($(GOOS),darwin)
ifeq ($(GOARCH),amd64)
	$(MAKE) build-cli-darwin
else
	$(MAKE) build-cli-darwin-arm
endif
else ifeq ($(GOOS),linux)
	$(MAKE) build-cli-linux
else
	@echo "Unsupported platform: $(GOOS)"
endif


build-cli-win:
	set CGO_ENABLED=1 && go build -o oadin.exe -ldflags="-s -w"  cmd/cli/main.go
	$(MAKE) trayapp

build-cli-darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64  go build -o oadin -ldflags="-s -w"  cmd/cli/main.go
	$(MAKE) trayapp

build-cli-darwin-arm:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64  go build -o oadin -ldflags="-s -w"  cmd/cli/main.go
	$(MAKE) trayapp

build-cli-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64  go build -o oadin -ldflags="-s -w"  cmd/cli/main.go
	$(MAKE) trayapp

trayapp:
ifeq ($(GOOS),windows)
	go build -o oadin-tray.exe trayapp/main.go
else
	go build -o oadin-tray trayapp/main.go
endif

build-for-ci:
ifeq ($(GOOS),windows)
	go build -o oadin-tray.exe trayapp/main.go
else
	go build -o oadin-tray trayapp/main.go
endif
	@echo "CI build completed"

copy-win-artifacts:
	copy /Y oadin.exe installer\win\
	copy /Y oadin-tray.exe installer\win\

copy-mac-artifacts:
	cp oadin installer/mac/
	cp oadin-tray installer/mac/
	chmod +x installer/mac/create-app.sh

build-mac-app: copy-mac-artifacts
	cd installer/mac && ./create-app.sh

build-win-installer: build-cli-win copy-win-artifacts
	cd installer\win && makensis oadin.nsi

build-mac-installer: build-cli-darwin build-mac-app

.PHONY: build-all build-cli-win build-cli-darwin build-cli-darwin-arm build-cli-linux trayapp build-for-ci copy-win-artifacts copy-mac-artifacts build-mac-app build-win-installer build-mac-installer
build-mac-installer: build-cli-darwin build-mac-app

build-for-pipeline:
	go build -o oadin-tray.exe trayapp/main.go
	@echo "Pipeline build completed with trayapp"


ensure-trayapp:
ifeq ($(GOOS),windows)
	@if not exist oadin-tray.exe ( \
		echo Building missing oadin-tray.exe... && \
		go build -o oadin-tray.exe trayapp/main.go \
	) else ( \
		echo oadin-tray.exe already exists \
	)
else
	@if [ ! -f oadin-tray ]; then \
		echo "Building missing oadin-tray..."; \
		go build -o oadin-tray trayapp/main.go; \
	else \
		echo "oadin-tray already exists"; \
	fi
endif


prepare-win-build:
	@echo "Preparing Windows build for CI..."
	go build -o oadin-tray.exe trayapp/main.go
	@echo "oadin-tray.exe built successfully"
	@if exist oadin.exe echo "oadin.exe found" else echo "Warning: oadin.exe not found"
	@if exist oadin-tray.exe echo "oadin-tray.exe found" else echo "Warning: oadin-tray.exe not found"

# 专门为流水线设计 - 无条件构建 trayapp
force-build-tray:
	@echo "Force building trayapp for pipeline..."
	go build -o oadin-tray.exe trayapp/main.go
	@echo "✅ oadin-tray.exe built successfully"

# 验证构建结果
verify-build:
	@echo "Verifying build artifacts..."
	@if exist oadin.exe ( echo "✅ oadin.exe found" ) else ( echo "❌ oadin.exe missing" && exit 1 )
	@if exist oadin-tray.exe ( echo "✅ oadin-tray.exe found" ) else ( echo "❌ oadin-tray.exe missing" && exit 1 )
	@echo "All artifacts verified successfully"

.PHONY: build-all build-cli-win build-cli-darwin build-cli-darwin-arm build-cli-linux trayapp build-for-ci copy-win-artifacts copy-mac-artifacts build-mac-app build-win-installer build-mac-installer ensure-trayapp prepare-win-build force-build-tray verify-build
