#!/bin/bash
set -euo pipefail

# Create AppImage for NetMgr
BINARY_PATH="${1:-build/netmgr}"
VERSION="${2:-$(git describe --tags --always)}"

if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ Binary not found: $BINARY_PATH"
    exit 1
fi

echo "📦 Creating AppImage for NetMgr $VERSION"

# Download AppImage tools
if [ ! -f "appimagetool-x86_64.AppImage" ]; then
    echo "⬇️  Downloading AppImage tools..."
    wget -q https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage
    chmod +x appimagetool-x86_64.AppImage
fi

# Create AppDir structure
APPDIR="NetMgr.AppDir"
rm -rf "$APPDIR"
mkdir -p "$APPDIR"/{usr/bin,usr/share/{applications,icons/hicolor/256x256/apps}}

# Copy binary
cp "$BINARY_PATH" "$APPDIR/usr/bin/netmgr"

# Create desktop file
cat > "$APPDIR/usr/share/applications/netmgr.desktop" << 'EOF'
[Desktop Entry]
Type=Application
Name=NetMgr
Comment=Cross-platform network management tool
Exec=netmgr
Icon=netmgr
Categories=Network;System;
Terminal=true
StartupNotify=true
EOF

# Create AppRun
cat > "$APPDIR/AppRun" << 'EOF'
#!/bin/bash
HERE="$(dirname "$(readlink -f "${0}")")"
export PATH="${HERE}/usr/bin:${PATH}"
exec "${HERE}/usr/bin/netmgr" "$@"
EOF
chmod +x "$APPDIR/AppRun"

# Copy desktop file to root
cp "$APPDIR/usr/share/applications/netmgr.desktop" "$APPDIR/"

# Create simple icon (you can replace with actual icon)
if command -v convert >/dev/null 2>&1; then
    convert -size 256x256 -background blue -fill white -gravity center \
        -pointsize 48 label:"NetMgr" "$APPDIR/netmgr.png"
    cp "$APPDIR/netmgr.png" "$APPDIR/usr/share/icons/hicolor/256x256/apps/"
else
    echo "⚠️  ImageMagick not found, skipping icon creation"
    touch "$APPDIR/netmgr.png"
fi

# Build AppImage
echo "🔨 Building AppImage..."
./appimagetool-x86_64.AppImage "$APPDIR" "netmgr-$VERSION-x86_64.AppImage"

echo "✅ AppImage created: netmgr-$VERSION-x86_64.AppImage"
