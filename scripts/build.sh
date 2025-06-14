#!/bin/bash
set -euo pipefail

# NetMgr Build Script
# Usage: ./scripts/build.sh [debug|release] [target]

BUILD_TYPE="${1:-release}"
TARGET="${2:-native}"

echo "🔨 Building NetMgr ($BUILD_TYPE mode, $TARGET target)"

# Create build directory
BUILD_DIR="build-$BUILD_TYPE"
mkdir -p "$BUILD_DIR"

# Configure CMake based on target
case "$TARGET" in
    "native")
        CMAKE_ARGS="-DCMAKE_BUILD_TYPE=${BUILD_TYPE^}"
        ;;
    "arm64")
        CMAKE_ARGS="-DCMAKE_BUILD_TYPE=${BUILD_TYPE^} -DCMAKE_TOOLCHAIN_FILE=cmake/aarch64-toolchain.cmake"
        ;;
    "windows")
        CMAKE_ARGS="-DCMAKE_BUILD_TYPE=${BUILD_TYPE^} -DCMAKE_TOOLCHAIN_FILE=cmake/mingw-toolchain.cmake"
        ;;
    *)
        echo "❌ Unknown target: $TARGET"
        exit 1
        ;;
esac

# Configure and build
echo "⚙️  Configuring CMake..."
cmake -B "$BUILD_DIR" -G Ninja $CMAKE_ARGS

echo "🏗️  Building..."
cmake --build "$BUILD_DIR" --config "${BUILD_TYPE^}"

# Run tests if available
if [ -f "$BUILD_DIR/test_netmgr" ]; then
    echo "🧪 Running tests..."
    cd "$BUILD_DIR" && ctest --output-on-failure
    cd ..
fi

echo "✅ Build complete! Binary: $BUILD_DIR/netmgr"

# Create packages for release builds
if [ "$BUILD_TYPE" = "release" ] && [ "$TARGET" = "native" ]; then
    echo "📦 Creating packages..."
    
    # Debian package
    if command -v dpkg-deb >/dev/null 2>&1; then
        make -C "$BUILD_DIR" package
        echo "✅ Debian package created"
    fi
    
    # AppImage (Linux only)
    if [ "$(uname)" = "Linux" ] && command -v wget >/dev/null 2>&1; then
        ./scripts/create-appimage.sh "$BUILD_DIR/netmgr"
        echo "✅ AppImage created"
    fi
fi

echo "🎉 All done!"
