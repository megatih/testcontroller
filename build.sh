#!/bin/bash

# SDL3 Gamepad Test Controller - macOS Build Script
# This script builds a complete macOS application bundle (.app) and creates a DMG for distribution

set -e  # Exit on error

# Configuration
APP_NAME="TestController"
APP_VERSION="1.0.0"
BUNDLE_ID="com.megatih.testcontroller"
BINARY_NAME="testcontroller"
DMG_NAME="${APP_NAME}-${APP_VERSION}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directories
BUILD_DIR="build"
APP_BUNDLE="${BUILD_DIR}/${APP_NAME}.app"
CONTENTS_DIR="${APP_BUNDLE}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"
FRAMEWORKS_DIR="${CONTENTS_DIR}/Frameworks"
DMG_DIR="${BUILD_DIR}/dmg"

echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  SDL3 Gamepad Test Controller - macOS Build Script            ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Step 1: Clean previous build
echo -e "${YELLOW}[1/7]${NC} Cleaning previous build..."
rm -rf "${BUILD_DIR}"
mkdir -p "${BUILD_DIR}"
echo -e "${GREEN}✓${NC} Clean complete"
echo ""

# Step 2: Build the Go binary
echo -e "${YELLOW}[2/7]${NC} Building Go binary..."
echo -e "      Compiling for macOS (arm64 + amd64 universal binary)..."

# Build for both architectures
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o "${BUILD_DIR}/${BINARY_NAME}-arm64" \
    -ldflags "-s -w" \
    .

CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o "${BUILD_DIR}/${BINARY_NAME}-amd64" \
    -ldflags "-s -w" \
    .

# Create universal binary
lipo -create \
    "${BUILD_DIR}/${BINARY_NAME}-arm64" \
    "${BUILD_DIR}/${BINARY_NAME}-amd64" \
    -output "${BUILD_DIR}/${BINARY_NAME}"

# Clean up architecture-specific binaries
rm -f "${BUILD_DIR}/${BINARY_NAME}-arm64" "${BUILD_DIR}/${BINARY_NAME}-amd64"

echo -e "${GREEN}✓${NC} Binary built successfully"
echo ""

# Step 3: Create app bundle structure
echo -e "${YELLOW}[3/7]${NC} Creating application bundle structure..."
mkdir -p "${MACOS_DIR}"
mkdir -p "${RESOURCES_DIR}"
mkdir -p "${FRAMEWORKS_DIR}"
echo -e "${GREEN}✓${NC} Bundle structure created"
echo ""

# Step 4: Copy executable and create launcher
echo -e "${YELLOW}[4/7]${NC} Creating launcher and copying executable..."

# Copy the actual binary with a different name
cp "${BUILD_DIR}/${BINARY_NAME}" "${MACOS_DIR}/${BINARY_NAME}-bin"
chmod +x "${MACOS_DIR}/${BINARY_NAME}-bin"

# Create a launcher script that sets the working directory to Resources
cat > "${MACOS_DIR}/${BINARY_NAME}" << 'LAUNCHER_EOF'
#!/bin/bash
# Launcher script for TestController
# Sets working directory to Resources so assets can be found

# Get the directory containing this script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Change to Resources directory (one level up, then into Resources)
cd "${DIR}/../Resources"

# Set DYLD_LIBRARY_PATH to find bundled SDL3
export DYLD_LIBRARY_PATH="${DIR}/../Frameworks:${DYLD_LIBRARY_PATH}"

# Run the actual binary
exec "${DIR}/testcontroller-bin" "$@"
LAUNCHER_EOF

chmod +x "${MACOS_DIR}/${BINARY_NAME}"
echo -e "${GREEN}✓${NC} Launcher created and executable copied"
echo ""

# Step 5: Copy resources and create app icon
echo -e "${YELLOW}[5/7]${NC} Copying resources and creating app icon..."
mkdir -p "${RESOURCES_DIR}/assets"

# Copy PNG assets
cp gamepadutils/assets/*.png "${RESOURCES_DIR}/assets/"
echo -e "      Copied $(ls gamepadutils/assets/*.png | wc -l | tr -d ' ') PNG files"

# Copy gamecontroller database
cp gamepadutils/assets/gamecontrollerdb.txt "${RESOURCES_DIR}/assets/"
echo -e "      Copied gamecontrollerdb.txt"

# Create app icon from PNG
if [ -f "testcontroller_icon_1024.png" ]; then
    echo -e "      Creating app icon..."

    # Create iconset directory
    ICONSET_DIR="${BUILD_DIR}/AppIcon.iconset"
    mkdir -p "${ICONSET_DIR}"

    # Generate all required icon sizes for macOS
    sips -z 16 16     testcontroller_icon_1024.png --out "${ICONSET_DIR}/icon_16x16.png" > /dev/null 2>&1
    sips -z 32 32     testcontroller_icon_1024.png --out "${ICONSET_DIR}/icon_16x16@2x.png" > /dev/null 2>&1
    sips -z 32 32     testcontroller_icon_1024.png --out "${ICONSET_DIR}/icon_32x32.png" > /dev/null 2>&1
    sips -z 64 64     testcontroller_icon_1024.png --out "${ICONSET_DIR}/icon_32x32@2x.png" > /dev/null 2>&1
    sips -z 128 128   testcontroller_icon_1024.png --out "${ICONSET_DIR}/icon_128x128.png" > /dev/null 2>&1
    sips -z 256 256   testcontroller_icon_1024.png --out "${ICONSET_DIR}/icon_128x128@2x.png" > /dev/null 2>&1
    sips -z 256 256   testcontroller_icon_1024.png --out "${ICONSET_DIR}/icon_256x256.png" > /dev/null 2>&1
    sips -z 512 512   testcontroller_icon_1024.png --out "${ICONSET_DIR}/icon_256x256@2x.png" > /dev/null 2>&1
    sips -z 512 512   testcontroller_icon_1024.png --out "${ICONSET_DIR}/icon_512x512.png" > /dev/null 2>&1
    sips -z 1024 1024 testcontroller_icon_1024.png --out "${ICONSET_DIR}/icon_512x512@2x.png" > /dev/null 2>&1

    # Convert iconset to icns
    iconutil -c icns "${ICONSET_DIR}" -o "${RESOURCES_DIR}/AppIcon.icns"

    # Clean up iconset directory
    rm -rf "${ICONSET_DIR}"

    echo -e "      Created AppIcon.icns"
else
    echo -e "      ${YELLOW}⚠${NC} Icon file testcontroller_icon_1024.png not found, skipping icon creation"
fi

echo -e "${GREEN}✓${NC} Resources copied"
echo ""

# Step 6: Copy SDL3 library
echo -e "${YELLOW}[6/7]${NC} Bundling SDL3 framework..."

# Check for SDL3 in common locations
SDL3_LIB=""
if [ -f "/usr/local/lib/libSDL3.dylib" ]; then
    SDL3_LIB="/usr/local/lib/libSDL3.dylib"
elif [ -f "/opt/homebrew/lib/libSDL3.dylib" ]; then
    SDL3_LIB="/opt/homebrew/lib/libSDL3.dylib"
else
    echo -e "${RED}✗${NC} SDL3 library not found!"
    echo -e "  Please install SDL3 via Homebrew: ${BLUE}brew install sdl3${NC}"
    exit 1
fi

# Copy SDL3 library
cp "${SDL3_LIB}" "${FRAMEWORKS_DIR}/"

# Update library paths to use @executable_path
install_name_tool -id "@executable_path/../Frameworks/libSDL3.dylib" \
    "${FRAMEWORKS_DIR}/libSDL3.dylib"

# Update binary to look for SDL3 in Frameworks
install_name_tool -change "${SDL3_LIB}" \
    "@executable_path/../Frameworks/libSDL3.dylib" \
    "${MACOS_DIR}/${BINARY_NAME}-bin"

# Also handle /usr/local/lib path if that's what was used during build
if [ "${SDL3_LIB}" != "/usr/local/lib/libSDL3.dylib" ]; then
    install_name_tool -change "/usr/local/lib/libSDL3.dylib" \
        "@executable_path/../Frameworks/libSDL3.dylib" \
        "${MACOS_DIR}/${BINARY_NAME}-bin" 2>/dev/null || true
fi

echo -e "${GREEN}✓${NC} SDL3 framework bundled"
echo ""

# Step 7: Create Info.plist
echo -e "${YELLOW}[7/7]${NC} Creating Info.plist..."

cat > "${CONTENTS_DIR}/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleExecutable</key>
    <string>${BINARY_NAME}</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>${BUNDLE_ID}</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${APP_VERSION}</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSHumanReadableCopyright</key>
    <string>Copyright © 2024. All rights reserved.</string>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.utilities</string>
</dict>
</plist>
EOF

echo -e "${GREEN}✓${NC} Info.plist created"
echo ""

# Step 8: Create DMG
echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Creating DMG for Distribution                                ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""

echo -e "${YELLOW}[8/8]${NC} Creating DMG file..."

# Create DMG staging directory
mkdir -p "${DMG_DIR}"
cp -R "${APP_BUNDLE}" "${DMG_DIR}/"

# Create Applications symlink for easy installation
ln -s /Applications "${DMG_DIR}/Applications"

# Create README in DMG
cat > "${DMG_DIR}/README.txt" << EOF
SDL3 Gamepad Test Controller v${APP_VERSION}

INSTALLATION:
1. Drag ${APP_NAME}.app to the Applications folder
2. Launch from Applications or Spotlight

USAGE:
Run the application and connect your gamepad to test buttons, axes, and sensors.

For more information, visit:
https://github.com/megatih/testcontroller

Built with SDL3 and Go.
EOF

# Remove old DMG if it exists
rm -f "${BUILD_DIR}/${DMG_NAME}.dmg"

# Create DMG
echo -e "      Creating disk image..."
hdiutil create -volname "${APP_NAME}" \
    -srcfolder "${DMG_DIR}" \
    -ov \
    -format UDZO \
    "${BUILD_DIR}/${DMG_NAME}.dmg"

echo -e "${GREEN}✓${NC} DMG created: ${DMG_NAME}.dmg"
echo ""

# Print summary
echo -e "${GREEN}╔════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Build Complete!                                               ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "📦 Application Bundle: ${BLUE}${APP_BUNDLE}${NC}"
echo -e "💿 Distribution DMG:   ${BLUE}${BUILD_DIR}/${DMG_NAME}.dmg${NC}"
echo ""
echo -e "App Bundle Size:  $(du -sh "${APP_BUNDLE}" | cut -f1)"
echo -e "DMG Size:         $(du -sh "${BUILD_DIR}/${DMG_NAME}.dmg" | cut -f1)"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "  • Test the app: ${BLUE}open \"${APP_BUNDLE}\"${NC}"
echo -e "  • Mount the DMG: ${BLUE}open \"${BUILD_DIR}/${DMG_NAME}.dmg\"${NC}"
echo -e "  • Distribute: Share ${DMG_NAME}.dmg with users"
echo ""
