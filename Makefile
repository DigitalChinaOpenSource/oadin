# 先获取 GOOS 和 GOARCH
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
SQLITE_VEC_DIR ?= $(abspath internal/datastore/sqlite/sqlite-vec)

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
	CGO_ENABLED=1 CGO_CFLAGS=-I$(SQLITE_VEC_DIR) GOOS=windows GOARCH=amd64 go build -o oadin.exe -ldflags="-s -w" cmd/cli/main.go

build-cli-darwin:
	CGO_ENABLED=1 CGO_CFLAGS=-I$(SQLITE_VEC_DIR) GOOS=darwin GOARCH=amd64 go build -o oadin -ldflags="-s -w" cmd/cli/main.go

build-cli-darwin-arm:
	CGO_ENABLED=1 CGO_CFLAGS=-I$(SQLITE_VEC_DIR) GOOS=darwin GOARCH=arm64 go build -o oadin -ldflags="-s -w" cmd/cli/main.go

build-cli-linux:
	CGO_ENABLED=1 CGO_CFLAGS=-I$(SQLITE_VEC_DIR) GOOS=linux GOARCH=amd64 go build -o oadin -ldflags="-s -w" cmd/cli/main.go

build-dll-win:
	CGO_ENABLED=1 CGO_CFLAGS=-I$(SQLITE_VEC_DIR) GOOS=windows GOARCH=amd64 go build -o OadinChecker.dll -buildmode=c-shared checker/OadinChecker.go

build-dll-darwin:
	CGO_ENABLED=1 CGO_CFLAGS=-I$(SQLITE_VEC_DIR) GOOS=darwin GOARCH=amd64 go build -o OadinChecker.dylib -buildmode=c-shared checker/OadinChecker.go

build-dll-darwin-arm:
	CGO_ENABLED=1 CGO_CFLAGS=-I$(SQLITE_VEC_DIR) GOOS=darwin GOARCH=arm64 go build -o OadinChecker.dylib -buildmode=c-shared checker/OadinChecker.go

build-dll-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o OadinChecker.so -buildmode=c-shared checker/OadinChecker.go