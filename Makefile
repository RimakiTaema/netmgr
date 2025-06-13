# Makefile for cross-compiling netmgr

# Binary name
BINARY_NAME=netmgr

# Build directory
BUILD_DIR=build

# Version
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Platforms to build for
PLATFORMS=linux/amd64 linux/386 linux/arm linux/arm64 windows/amd64 windows/386 darwin/amd64 darwin/arm64

.PHONY: all clean build-all $(PLATFORMS)

all: build-all

clean:
	rm -rf $(BUILD_DIR)

build-all: $(PLATFORMS)

$(PLATFORMS):
	$(eval PLATFORM_SPLIT := $(subst /, ,$@))
	$(eval GOOS := $(word 1, $(PLATFORM_SPLIT)))
	$(eval GOARCH := $(word 2, $(PLATFORM_SPLIT)))
	$(eval OUTPUT_NAME := $(BINARY_NAME))
	@if [ "$(GOOS)" = "windows" ]; then \
		OUTPUT_NAME=$(BINARY_NAME).exe; \
	fi
	@echo "Building for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)/$(GOOS)-$(GOARCH)
	@GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(GOOS)-$(GOARCH)/$(OUTPUT_NAME) ./cmd/netmgr
	@echo "Done!"

# Build for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/netmgr

# Install to system
install: build
	@if [ "$(shell uname)" = "Linux" ] || [ "$(shell uname)" = "Darwin" ]; then \
		sudo cp $(BINARY_NAME) /usr/local/bin/; \
		echo "Installed to /usr/local/bin/$(BINARY_NAME)"; \
	else \
		echo "Installation not supported on this platform"; \
	fi

# Run tests
test:
	go test -v ./...
