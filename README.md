# SDL3 Gamepad Test Controller (Go Port)

A complete pure Go port of SDL3's `testcontroller` application, providing comprehensive gamepad testing, visualization, and mapping capabilities using the [go-sdl3](https://github.com/Zyko0/go-sdl3) bindings.

![License](https://img.shields.io/badge/license-zlib-blue)
![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?logo=go)
![Platform](https://img.shields.io/badge/platform-linux%20%7C%20windows%20%7C%20macOS-lightgrey)

## Features

### Core Functionality
- **Real-time Gamepad Visualization** - Interactive visual representation of all gamepad inputs
- **Button & Axis Testing** - Live feedback for all buttons, axes, triggers, and D-pad
- **Touchpad Support** - PS5 DualSense touchpad visualization and testing
- **IMU/Gyroscope Display** - 3D visualization of gyroscope and accelerometer data
- **Virtual Gamepad** - Software gamepad simulation for testing without physical hardware
- **Custom Mapping Creation** - Interactive guided binding system for creating controller mappings
- **PS5-Specific Features** - Adaptive trigger effects, LED control, and audio routing

### Advanced Features
- **Gyro Calibration** - Multi-phase calibration system for drift correction
  - Noise profiling phase
  - Drift profiling phase
  - Real-time drift compensation
- **3D Gyro Visualization** - Wireframe gizmo showing real-time orientation
- **Sensor Rate Detection** - Reports both advertised and estimated sensor update rates
- **Multiple Controller Support** - Test multiple gamepads simultaneously
- **Gamepad Type Selection** - Support for Xbox, PlayStation, Nintendo, and generic layouts
- **Mapping Database** - Ships with SDL's `gamecontrollerdb.txt` for wide hardware support

## Screenshots

The application displays:
- **Left Panel**: Gamepad element list with live state indicators
- **Center Panel**: Visual gamepad representation with button/axis highlighting
- **Right Panel**: Raw joystick data or IMU/gyro information
- **Status Bar**: Controller info, battery level, connection type

## Prerequisites

### Runtime Requirements
- **SDL3** shared library installed on your system
  - Linux: `libSDL3.so.0`
  - Windows: `SDL3.dll`
  - macOS: `libSDL3.dylib`

### Build Requirements
- Go 1.21 or later
- SDL3 development headers (for testing, though go-sdl3 is pure Go with no CGo)

### Installing SDL3

#### Linux (Ubuntu/Debian)
```bash
# Install from SDL3 releases or build from source
# https://github.com/libsdl-org/SDL/releases
```

#### macOS (Homebrew)
```bash
brew install sdl3
```

#### Windows
Download SDL3 runtime from [SDL Releases](https://github.com/libsdl-org/SDL/releases) and place `SDL3.dll` in your PATH or application directory.

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/testcontroller.git
cd testcontroller

# Download dependencies
go mod download

# Build the application
go build -o testcontroller .
```

## Usage

### Basic Usage
```bash
# Run with physical gamepad
./testcontroller

# Run with virtual gamepad (automatically creates one)
./testcontroller --virtual
```

### Keyboard Controls

#### General
- `ESC` - Quit application
- `Space` - Cycle between front/back gamepad view
- `A` - Create/remove virtual gamepad
- `C` - Clear current gamepad mapping
- `X` - Copy current mapping to clipboard (when available)
- `V` - Paste mapping from clipboard (when available)

#### Binding Mode
- `B` - Enter/exit binding mode (setup custom mapping)
- `N` - Edit controller name in binding mode
- `Backspace` - Delete characters when editing name
- `Return/Enter` - Commit element binding or finish name editing
- `Left Arrow` - Previous element in binding sequence
- `Right Arrow` - Next element in binding sequence (skip current)

#### PS5 DualSense Controls
- `T` - Cycle trigger effects (Off → Pulse → Stiff → Vibrate)
- `M` - Cycle audio routing (Speakers → Headphones → Both)
- **Left Stick** - Controls LED color in real-time

#### IMU/Gyroscope
- `R` - Reset gyroscope orientation
- Click **Calibrate** button - Start drift calibration process

### Virtual Gamepad Interaction
When a virtual gamepad is active, you can interact with the on-screen controller using your mouse:
- **Click buttons** - Activate/deactivate
- **Drag sticks** - Move analog sticks
- **Drag triggers** - Pull triggers
- **Click touchpad** - Simulate touch input

## Project Structure

```
testcontroller/
├── main.go              # SDL initialization, main loop, event handling
├── controller.go        # Controller state management, PS5 effects
├── binding.go          # Mapping creation and binding logic
├── quaternion.go       # Quaternion math for gyro visualization
├── imu.go              # IMU calibration state machine
├── virtual.go          # Virtual gamepad implementation
├── draw.go             # Rendering utilities
├── constants.go        # Layout constants and helpers
├── gamepadutils/       # UI widget library
│   ├── types.go        # Enums and type definitions
│   ├── image.go        # Gamepad visual representation
│   ├── display.go      # Element list display
│   ├── gyro_display.go # IMU/gyro visualization
│   ├── joystick_display.go # Raw joystick element display
│   ├── button.go       # Button widget
│   ├── type_display.go # Gamepad type picker
│   ├── mapping.go      # Mapping CSV parser/generator
│   ├── wireframe.go    # 3D rendering utilities
│   └── embed.go        # Asset embedding
└── assets/             # Embedded resources
    ├── *.bmp           # Gamepad visualization images (16 files)
    └── gamecontrollerdb.txt  # SDL mapping database
```

## Architecture

### Key Design Decisions

1. **Pure Go Implementation** - Uses `github.com/Zyko0/go-sdl3` pure Go bindings (no CGo)
2. **Direct Port Fidelity** - Maintains structure and behavior of original C implementation
3. **Embedded Assets** - All BMP images and mapping database embedded using `go:embed`
4. **Package-Level State** - Matches C's static globals for direct translation accuracy
5. **Method-Based SDL API** - go-sdl3 uses methods on types (e.g., `button.GamepadStringForButton()`)

### Event Flow
```
main()
  → sdl.LoadLibrary()
  → appInit()
  → Event Loop
      → sdl.PollEvent()
      → handleEvent()
      → iterate() (update & render)
  → appQuit()
  → sdl.CloseLibrary()
```

### Display Components
- **GamepadImage** - Central visual gamepad with button/axis positions
- **GamepadDisplay** - Left panel showing element names and values
- **GyroDisplay** - 3D gyro visualization with calibration UI
- **JoystickDisplay** - Raw joystick button/axis/hat data
- **GamepadTypeDisplay** - Controller type selection interface

## Technical Details

### SDL3 API Differences from C

The go-sdl3 binding has some key differences from the C API:

| C API | Go API |
|-------|--------|
| `SDL_GetGamepadStringForButton(btn)` | `btn.GamepadStringForButton()` |
| `SDL_CreateRenderer(win, name)` | `win.CreateRenderer(name)` |
| `SDL_OpenGamepad(id)` | `id.OpenGamepad()` |
| `renderer.DrawColor()` | Returns `(sdl.Color, error)` not `(r,g,b,a,err)` |
| `SDL_NS_PER_SECOND` | Not defined - use local constant |
| `SDL_JOYSTICK_AXIS_MAX` | Not defined - use local constant (32767) |

### IMU Calibration Algorithm

The gyro calibration uses a two-phase approach:

1. **Noise Profiling** (5 seconds)
   - Measures accelerometer variance while stationary
   - Establishes noise threshold
   - Validates controller is on stable surface

2. **Drift Profiling** (30 seconds)
   - Samples gyroscope readings below noise threshold
   - Calculates average drift per axis
   - Applies drift compensation to quaternion integration

### PS5 DualSense Effects

The application supports PS5-specific features via `SendEffect()`:
- **Trigger Effects**: Pulse, Stiffness, and Vibration modes
- **LED Control**: RGB LED controlled by left analog stick position
- **Audio Routing**: Switch between speakers, headphones, or both

## Building from Source

### Quick Build

```bash
# Clone repository
git clone https://github.com/yourusername/testcontroller.git
cd testcontroller

# Verify dependencies
go mod tidy

# Build optimized binary
go build -ldflags="-s -w" -o testcontroller .

# Run tests (if any)
go test ./...

# Check code quality
go vet ./...
```

### macOS Application Bundle

For macOS, you can build a distributable .app bundle and DMG installer:

#### Quick Reference

```bash
# Build the app
./build.sh
```

**Output:**
```
build/
├── TestController.app/              ← macOS application bundle
└── TestController-1.0.0.dmg         ← Distributable disk image
```

**Test:**
```bash
open build/TestController.app        # Test the app
open build/TestController-1.0.0.dmg  # Test the DMG
```

#### What's Included in the .app Bundle

✅ Universal binary (Apple Silicon + Intel)
✅ Self-contained SDL3 library (no Homebrew needed)
✅ Custom app icon (from testcontroller_icon_1024.png)
✅ All PNG assets for gamepad graphics
✅ Gamepad controller database
✅ Launcher script for proper resource loading
✅ Info.plist with app metadata

#### Distribution

Share the DMG file with users:
1. They open TestController-1.0.0.dmg
2. Drag TestController.app to Applications folder
3. Launch from Applications or Spotlight

**No installation of SDL3 or other dependencies required!**

#### Build Prerequisites (macOS)

Before building the app bundle, ensure you have:

1. **Go** (1.21 or later)
   ```bash
   brew install go
   ```

2. **SDL3** installed via Homebrew
   ```bash
   brew install sdl3
   ```

3. **Xcode Command Line Tools** (for `lipo`, `install_name_tool`, `hdiutil`)
   ```bash
   xcode-select --install
   ```

#### Build Process

The `build.sh` script:
1. ✅ Cleans previous builds
2. ✅ Compiles universal binary (Apple Silicon + Intel)
3. ✅ Creates `.app` bundle structure
4. ✅ Copies executable and creates launcher
5. ✅ Copies all resources (PNG assets, gamecontroller database)
6. ✅ Creates app icon from `testcontroller_icon_1024.png` (if present)
7. ✅ Bundles SDL3 library with proper linking
8. ✅ Generates `Info.plist` with icon reference
9. ✅ Creates distributable DMG file

#### App Icon

Place a 1024x1024 PNG icon file named `testcontroller_icon_1024.png` in the project root.
The build script will automatically:
- Generate all required icon sizes (16x16 to 512x512 @2x)
- Create a proper macOS `.icns` file
- Include it in the app bundle
- Reference it in `Info.plist`

If the icon file is not present, the build will continue without an icon.

#### App Bundle Structure

```
TestController.app/
└── Contents/
    ├── Info.plist
    ├── MacOS/
    │   ├── testcontroller        # Launcher script
    │   └── testcontroller-bin    # Actual binary
    ├── Resources/
    │   ├── AppIcon.icns          # Application icon
    │   └── assets/
    │       ├── *.png             # Gamepad graphics
    │       └── gamecontrollerdb.txt
    └── Frameworks/
        └── libSDL3.dylib         # Bundled SDL3
```

#### Customization

You can customize the build by editing `build.sh` variables:

```bash
APP_NAME="TestController"                      # Application name
APP_VERSION="1.0.0"                            # Version number
BUNDLE_ID="com.megatih.testcontroller"         # Bundle identifier
DMG_NAME="${APP_NAME}-${APP_VERSION}"          # DMG filename
```

#### Code Signing (Optional)

For public distribution, you may want to code sign:

```bash
codesign --deep --force --sign "Developer ID Application: Your Name" \
    build/TestController.app
```

Then notarize with Apple for Gatekeeper:
```bash
xcrun notarytool submit build/TestController-1.0.0.dmg \
    --apple-id your@email.com \
    --password "app-specific-password" \
    --team-id TEAMID
```

## Troubleshooting

### SDL3 Library Not Found
```
Error: Failed to load SDL3 library
```
**Solution**: Install SDL3 runtime library for your platform (see Prerequisites)

### Window Doesn't Appear
**Solution**: Check if `$DISPLAY` is set (Linux/X11) or if you have graphical environment access

### Gamepad Not Detected
1. Verify gamepad is connected: `ls /dev/input/` (Linux)
2. Check SDL3 can see it: Set `SDL_LOGGING=info` environment variable
3. Try with `--virtual` flag to test UI

### Virtual Gamepad Issues
**Solution**: Press `A` key to toggle virtual gamepad creation/removal

### Gyro Calibration Fails
**Solution**: Place controller on a completely flat, stable surface during calibration

## Development Notes

### Vendored go-sdl3 Fixes

This project vendors the go-sdl3 dependency and applies fixes to unimplemented methods.

#### Working Fixes (Applied)

1. **`Gamepad.TouchpadFinger()`** - `vendor/github.com/Zyko0/go-sdl3/sdl/methods.go:2708`
   - **Fix**: Removed `panic("not implemented")`
   - **Status**: ✅ Working - Touchpad finger tracking now functional

2. **`Gamepad.SensorData()`** - `vendor/github.com/Zyko0/go-sdl3/sdl/methods.go:2742`
   - **Fix**: Removed `panic("not implemented")`
   - **Status**: ✅ Working - Accelerometer and gyro data now accessible

#### Known Issues (Not Fixed)

3. **`JoystickID.JoystickGUIDForID()`** and **`Joystick.GUID()`**
   - **Issue**: GUID struct-return ABI incompatibility with purego
   - **Impact**: GUID display disabled in verbose logging (non-critical)
   - **Technical**: SDL3 returns `SDL_GUID` structs by value (16 bytes), but purego's `SyscallN` expects pointer returns, resulting in invalid memory access

#### What Works

- ✅ Touchpad finger position tracking
- ✅ Touchpad finger pressure detection
- ✅ Accelerometer sensor data (3-axis)
- ✅ Gyroscope sensor data (3-axis)
- ✅ All button/axis/hat inputs
- ✅ Controller connection/disconnection
- ❌ GUID display (not critical for functionality)

#### Updating go-sdl3

If you need to update go-sdl3 in the future:

1. Update go.mod version
2. Run `go mod vendor`
3. Re-apply these fixes to `vendor/github.com/Zyko0/go-sdl3/sdl/methods.go`:
   - Remove panic from `TouchpadFinger()` (line ~2708)
   - Remove panic from `SensorData()` (line ~2742)
   - Keep panic in GUID methods with explanation

#### Future Upstream Contribution

The `TouchpadFinger` and `SensorData` fixes should be submitted as a PR to [Zyko0/go-sdl3](https://github.com/Zyko0/go-sdl3).

The GUID issue requires deeper changes to handle struct-by-value returns in purego, which may need upstream purego library support or special handling in go-sdl3.

## Contributing

Contributions are welcome! This project aims to maintain parity with SDL3's official `testcontroller.c`. When contributing:

1. Keep code structure similar to C implementation for maintainability
2. Follow Go best practices and idioms
3. Test with multiple gamepad types
4. Update documentation for new features

## Credits

- **Original Implementation**: SDL3 test suite (`test/testcontroller.c`, `test/gamepadutils.c/h`)
- **SDL3 Library**: [libsdl.org](https://libsdl.org/)
- **go-sdl3 Bindings**: [Zyko0/go-sdl3](https://github.com/Zyko0/go-sdl3)
- **Gamepad Database**: [SDL_GameControllerDB](https://github.com/gabomdq/SDL_GameControllerDB)

## License

This project uses the same zlib license as SDL3:

```
Copyright (C) 1997-2026 Sam Lantinga <slouken@libsdl.org>

This software is provided 'as-is', without any express or implied
warranty. In no event will the authors be held liable for any damages
arising from the use of this software.

Permission is granted to anyone to use this software for any purpose,
including commercial applications, and to alter it and redistribute it
freely.
```

## Related Projects

- [SDL3](https://github.com/libsdl-org/SDL) - Simple DirectMedia Layer
- [go-sdl3](https://github.com/Zyko0/go-sdl3) - Pure Go SDL3 bindings
- [SDL_GameControllerDB](https://github.com/gabomdq/SDL_GameControllerDB) - Community gamepad mapping database

## Roadmap

- [ ] Add screenshot capability
- [ ] Export custom mappings to file
- [ ] Gamepad recording/playback
- [ ] Haptic/rumble pattern editor
- [ ] Multi-language support
- [ ] Configuration file for preferences

---

**Status**: ✅ Feature-complete port of SDL3 testcontroller
**Tested**: Linux (ARM64)
**Go Version**: 1.21+
**SDL3 Version**: 3.x
