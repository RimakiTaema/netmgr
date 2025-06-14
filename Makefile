# Makefile for cross-compiling netmgr in Rust

# Binary name
BINARY_NAME=netmgr

# Build directory
BUILD_DIR=target

# Version
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Platforms to build for
PLATFORMS=x86_64-unknown-linux-gnu \
          i686-unknown-linux-gnu \
          aarch64-unknown-linux-gnu \
          x86_64-pc-windows-gnu \
          i686-pc-windows-gnu \
          x86_64-apple-darwin \
          aarch64-apple-darwin

.PHONY: all clean build-all install test $(PLATFORMS)

all: build-all

clean:
	cargo clean

build-all: $(PLATFORMS)

$(PLATFORMS):
	@echo "Building for $@..."
	@rustup target add $@ 2>/dev/null || true
	@cargo build --release --target $@
	@echo "Done!"

# Build for current platform
build:
	cargo build --release

# Install to system
install: build
	@if [ "$(shell uname)" = "Linux" ] || [ "$(shell uname)" = "Darwin" ]; then \
		sudo cp target/release/$(BINARY_NAME) /usr/local/bin/; \
		echo "Installed to /usr/local/bin/$(BINARY_NAME)"; \
	else \
		echo "Installation not supported on this platform"; \
	fi

# Run tests
test:
	cargo test

# Format code
fmt:
	cargo fmt

# Check code
check:
	cargo check

# Clippy linting
clippy:
	cargo clippy -- -D warnings

# Build documentation
docs:
	cargo doc --open
