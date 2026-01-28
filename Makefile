# ============================================================
# Makefile for ash (Windows installer via Inno Setup on CI)
# ============================================================

BINARY      := ash
DIST        := dist
VERSION     ?= 2.1.0

LDFLAGS     := -s -w
BUILD_FLAGS := -trimpath -ldflags "$(LDFLAGS)"

.PHONY: tidy build dist-win clean

tidy:
	go mod tidy

# Build for current host
build:
	mkdir -p $(DIST)
	go build $(BUILD_FLAGS) -o $(DIST)/$(BINARY) .

dist-win: clean
	mkdir -p $(DIST)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o $(DIST)/$(BINARY)-windows-amd64.exe .

clean:
	rm -rf $(DIST)
