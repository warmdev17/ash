# Simple cross-platform builds and Windows global install
# Usage:
#   make tidy
#   make build           # build current OS/ARCH
#   make dist            # build all targets to ./dist/
#   make install-win     # (Windows, user) copy to %USERPROFILE%\bin and add PATH
#   make clean

MODULE := $(shell go list -m)
BINARY := ash
DIST   := dist

LDFLAGS := -s -w
BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

# Default: build for host
build:
	go build $(BUILD_FLAGS) -o $(DIST)/$(BINARY) .

tidy:
	go mod tidy

dist: clean
	mkdir -p $(DIST)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(DIST)/$(BINARY)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(DIST)/$(BINARY)-windows-arm64.exe .
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(DIST)/$(BINARY)-linux-amd64 .
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(DIST)/$(BINARY)-darwin-arm64 .

# Windows user-level install (no admin): adds %USERPROFILE%\bin to PATH if needed
install-win:
	powershell -ExecutionPolicy Bypass -File scripts/install_windows_user.ps1 "$(BINARY)" "$(DIST)"

clean:
	rm -rf $(DIST)
