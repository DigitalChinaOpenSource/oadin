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

build-dll-win:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o OadinChecker.dll -buildmode=c-shared checker/OadinChecker.go

build-dll-darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o OadinChecker.dylib -buildmode=c-shared checker/OadinChecker.go

build-dll-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o OadinChecker.so -buildmode=c-shared checker/OadinChecker.go

trayapp:
ifeq ($(GOOS),windows)
	go build -o oadin-tray.exe trayapp/main.go
else
	go build -o oadin-tray trayapp/main.go
endif

copy-win-artifacts:
	copy /Y oadin.exe installer\win\
	copy /Y oadin-tray.exe installer\win\
	copy /Y installer\win\preinstall.bat installer\win\
	copy /Y installer\win\postinstall.bat installer\win\
	copy /Y installer\win\start-oadin.bat installer\win\

copy-mac-artifacts:
	cp oadin installer/mac/
	cp oadin-tray installer/mac/
	chmod +x installer/mac/create-app.sh

build-mac-app: copy-mac-artifacts
	cd installer/mac && ./create-app.sh

build-win-installer: build-cli-win copy-win-artifacts
	cd installer\win && makensis oadin.nsi

build-mac-installer: build-cli-darwin build-mac-app

.PHONY: build-all build-cli-win trayapp copy-win-artifacts copy-mac-artifacts build-win-installer build-mac-installer build-mac-app
