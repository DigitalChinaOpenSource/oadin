

build-all:
ifeq ($(shell go env GOOS), windows)
	$(MAKE) build-cli-win build-dll-win
else ifeq ($(shell go env GOOS), darwin)
	$(MAKE) build-cli-darwin build-dll-darwin
else ifeq ($(shell go env GOOS), linux)
	$(MAKE) build-cli-linux build-dll-linux
else
	@echo "Unsupported platform: $(shell go env GOOS)"
endif


build-cli-win:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64  go build -o aog.exe -ldflags="-s -w"  cmd/cli/main.go

build-cli-darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64  go build -o aog -ldflags="-s -w"  cmd/cli/main.go

build-cli-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64  go build -o aog -ldflags="-s -w"  cmd/cli/main.go

build-dll-win:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o AogChecker.dll -buildmode=c-shared checker/AogChecker.go

build-dll-darwin:
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o AogChecker.dylib -buildmode=c-shared checker/AogChecker.go

build-dll-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o AogChecker.so -buildmode=c-shared checker/AogChecker.go