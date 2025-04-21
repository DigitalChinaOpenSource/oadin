# 先获取 GOOS 和 GOARCH
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

build-all:
ifeq ($(GOOS),windows)
	$(MAKE) build-cli-win build-dll-win
else ifeq ($(GOOS),darwin)
ifeq ($(GOARCH),amd64)
	$(MAKE) build-cli-darwin build-dll-darwin
else
	$(MAKE) build-cli-darwin-arm build-dll-darwin-arm
endif
else ifeq ($(GOOS),linux)
	$(MAKE) build-cli-linux build-dll-linux
else
	@echo "Unsupported platform: $(GOOS)"
endif


build-cli-win:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64  go build -o byze.exe -ldflags="-s -w"  cmd/cli/main.go

build-cli-darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64  go build -o byze -ldflags="-s -w"  cmd/cli/main.go

build-cli-darwin-arm:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64  go build -o byze -ldflags="-s -w"  cmd/cli/main.go

build-cli-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64  go build -o byze -ldflags="-s -w"  cmd/cli/main.go

build-dll-win:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o ByzeChecker.dll -buildmode=c-shared checker/ByzeChecker.go

build-dll-darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o ByzeChecker.dylib -buildmode=c-shared checker/ByzeChecker.go

build-dll-darwin-arm:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o ByzeChecker.dylib -buildmode=c-shared checker/ByzeChecker.go

build-dll-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o ByzeChecker.so -buildmode=c-shared checker/ByzeChecker.go