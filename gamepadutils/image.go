package gamepadutils

import (
	"math"
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"
)

// saveColor saves the current renderer draw color.
func saveColor(renderer *sdl.Renderer) sdl.Color {
	clr, _ := renderer.DrawColor()
	return clr
}

// restoreColor restores a previously saved draw color.
func restoreColor(renderer *sdl.Renderer, c sdl.Color) {
	renderer.SetDrawColor(c.R, c.G, c.B, c.A)
}

// SaveColor saves the current renderer draw color (exported version).
func SaveColor(renderer *sdl.Renderer) sdl.Color {
	return saveColor(renderer)
}

// RestoreColor restores a previously saved draw color (exported version).
func RestoreColor(renderer *sdl.Renderer, c sdl.Color) {
	restoreColor(renderer, c)
}

// Button element positions on the gamepad image (indexed by SDL_GamepadButton).
var buttonPositions = [20]struct{ X, Y int }{
	{413, 190}, // SDL_GAMEPAD_BUTTON_SOUTH
	{456, 156}, // SDL_GAMEPAD_BUTTON_EAST
	{372, 159}, // SDL_GAMEPAD_BUTTON_WEST
	{415, 127}, // SDL_GAMEPAD_BUTTON_NORTH
	{199, 157}, // SDL_GAMEPAD_BUTTON_BACK
	{257, 153}, // SDL_GAMEPAD_BUTTON_GUIDE
	{314, 157}, // SDL_GAMEPAD_BUTTON_START
	{98, 177},  // SDL_GAMEPAD_BUTTON_LEFT_STICK
	{331, 254}, // SDL_GAMEPAD_BUTTON_RIGHT_STICK
	{102, 65},  // SDL_GAMEPAD_BUTTON_LEFT_SHOULDER
	{421, 61},  // SDL_GAMEPAD_BUTTON_RIGHT_SHOULDER
	{179, 213}, // SDL_GAMEPAD_BUTTON_DPAD_UP
	{179, 274}, // SDL_GAMEPAD_BUTTON_DPAD_DOWN
	{141, 242}, // SDL_GAMEPAD_BUTTON_DPAD_LEFT
	{211, 242}, // SDL_GAMEPAD_BUTTON_DPAD_RIGHT
	{257, 199}, // SDL_GAMEPAD_BUTTON_MISC1
	{157, 160}, // SDL_GAMEPAD_BUTTON_RIGHT_PADDLE1
	{355, 160}, // SDL_GAMEPAD_BUTTON_LEFT_PADDLE1
	{157, 200}, // SDL_GAMEPAD_BUTTON_RIGHT_PADDLE2
	{355, 200}, // SDL_GAMEPAD_BUTTON_LEFT_PADDLE2
}

// Axis element positions on the gamepad image.
var axisPositions = [10]struct {
	X, Y  int
	Angle float64
}{
	{99, 178, 270.0},  // SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE
	{99, 178, 90.0},   // SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_POSITIVE
	{99, 178, 0.0},    // SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_NEGATIVE
	{99, 178, 180.0},  // SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_POSITIVE
	{331, 256, 270.0}, // SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_NEGATIVE
	{331, 256, 90.0},  // SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_POSITIVE
	{331, 256, 0.0},   // SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_NEGATIVE
	{331, 256, 180.0}, // SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_POSITIVE
	{116, 5, 180.0},   // SDL_GAMEPAD_ELEMENT_AXIS_LEFT_TRIGGER
	{400, 5, 180.0},   // SDL_GAMEPAD_ELEMENT_AXIS_RIGHT_TRIGGER
}

var touchpadArea = sdl.FRect{X: 148.0, Y: 20.0, W: 216.0, H: 118.0}

// GamepadTouchpadFinger represents a touchpad finger state.
type GamepadTouchpadFinger struct {
	Down     bool
	X, Y     float32
	Pressure float32
}

// GamepadImage renders a visual representation of a gamepad.
type GamepadImage struct {
	renderer     *sdl.Renderer
	frontTexture *sdl.Texture
	backTexture  *sdl.Texture

	faceABXYTexture *sdl.Texture
	faceAXBYTexture *sdl.Texture
	faceBAYXTexture *sdl.Texture
	faceSonyTexture *sdl.Texture

	connectionTexture [2]*sdl.Texture
	batteryTexture    [2]*sdl.Texture
	touchpadTexture   *sdl.Texture
	buttonTexture     *sdl.Texture
	axisTexture       *sdl.Texture

	gamepadWidth     float32
	gamepadHeight    float32
	faceWidth        float32
	faceHeight       float32
	connectionWidth  float32
	connectionHeight float32
	batteryWidth     float32
	batteryHeight    float32
	touchpadWidth    float32
	touchpadHeight   float32
	ButtonWidth      float32
	ButtonHeight     float32
	AxisWidth        float32
	AxisHeight       float32

	x, y            float32
	showingFront    bool
	showingTouchpad bool
	gamepadType     sdl.GamepadType
	eastLabel       sdl.GamepadButtonLabel
	displayMode     ControllerDisplayMode

	elements [SDL_GAMEPAD_ELEMENT_MAX]bool

	connectionState sdl.JoystickConnectionState
	batteryState    sdl.PowerState
	batteryPercent  int

	numFingers int
	fingers    []GamepadTouchpadFinger
}

// CreateGamepadImage creates a new gamepad image widget.
func CreateGamepadImage(renderer *sdl.Renderer) *GamepadImage {
	ctx := &GamepadImage{
		renderer:     renderer,
		showingFront: true,
	}

	ctx.frontTexture = CreateTextureFromPNG(renderer, gamepadFrontPNG)
	ctx.backTexture = CreateTextureFromPNG(renderer, gamepadBackPNG)
	if ctx.frontTexture != nil {
		ctx.gamepadWidth, ctx.gamepadHeight, _ = ctx.frontTexture.Size()
	}

	ctx.faceABXYTexture = CreateTextureFromPNG(renderer, gamepadFaceABXYPNG)
	ctx.faceAXBYTexture = CreateTextureFromPNG(renderer, gamepadFaceAXBYPNG)
	ctx.faceBAYXTexture = CreateTextureFromPNG(renderer, gamepadFaceBAYXPNG)
	ctx.faceSonyTexture = CreateTextureFromPNG(renderer, gamepadFaceSonyPNG)
	if ctx.faceABXYTexture != nil {
		ctx.faceWidth, ctx.faceHeight, _ = ctx.faceABXYTexture.Size()
	}

	ctx.connectionTexture[0] = CreateTextureFromPNG(renderer, gamepadWiredPNG)
	ctx.connectionTexture[1] = CreateTextureFromPNG(renderer, gamepadWirelessPNG)
	if ctx.connectionTexture[0] != nil {
		ctx.connectionWidth, ctx.connectionHeight, _ = ctx.connectionTexture[0].Size()
	}

	ctx.batteryTexture[0] = CreateTextureFromPNG(renderer, gamepadBatteryPNG)
	ctx.batteryTexture[1] = CreateTextureFromPNG(renderer, gamepadBatteryWiredPNG)
	if ctx.batteryTexture[0] != nil {
		ctx.batteryWidth, ctx.batteryHeight, _ = ctx.batteryTexture[0].Size()
	}

	ctx.touchpadTexture = CreateTextureFromPNG(renderer, gamepadTouchpadPNG)
	if ctx.touchpadTexture != nil {
		ctx.touchpadWidth, ctx.touchpadHeight, _ = ctx.touchpadTexture.Size()
	}

	ctx.buttonTexture = CreateTextureFromPNG(renderer, gamepadButtonPNG)
	if ctx.buttonTexture != nil {
		ctx.ButtonWidth, ctx.ButtonHeight, _ = ctx.buttonTexture.Size()
		ctx.buttonTexture.SetColorMod(10, 255, 21)
	}

	ctx.axisTexture = CreateTextureFromPNG(renderer, gamepadAxisPNG)
	if ctx.axisTexture != nil {
		ctx.AxisWidth, ctx.AxisHeight, _ = ctx.axisTexture.Size()
		ctx.axisTexture.SetColorMod(10, 255, 21)
	}

	return ctx
}

// SetPosition sets the top-left position of the gamepad image.
func (ctx *GamepadImage) SetPosition(x, y float32) {
	if ctx == nil {
		return
	}
	ctx.x = x
	ctx.y = y
}

// GetArea returns the bounding rectangle of the gamepad image.
func (ctx *GamepadImage) GetArea() sdl.FRect {
	if ctx == nil {
		return sdl.FRect{}
	}
	area := sdl.FRect{
		X: ctx.x,
		Y: ctx.y,
		W: ctx.gamepadWidth,
		H: ctx.gamepadHeight,
	}
	if ctx.showingTouchpad {
		area.H += ctx.touchpadHeight
	}
	return area
}

// GetTouchpadArea returns the touchpad input area in screen coordinates.
func (ctx *GamepadImage) GetTouchpadArea() sdl.FRect {
	if ctx == nil {
		return sdl.FRect{}
	}
	return sdl.FRect{
		X: ctx.x + (ctx.gamepadWidth-ctx.touchpadWidth)/2 + touchpadArea.X,
		Y: ctx.y + ctx.gamepadHeight + touchpadArea.Y,
		W: touchpadArea.W,
		H: touchpadArea.H,
	}
}

// SetShowingFront sets whether the front or back of the gamepad is shown.
func (ctx *GamepadImage) SetShowingFront(front bool) {
	if ctx == nil {
		return
	}
	ctx.showingFront = front
}

// GetType returns the current gamepad type.
func (ctx *GamepadImage) GetType() sdl.GamepadType {
	if ctx == nil {
		return sdl.GAMEPAD_TYPE_UNKNOWN
	}
	return ctx.gamepadType
}

// SetDisplayMode sets the display mode.
func (ctx *GamepadImage) SetDisplayMode(mode ControllerDisplayMode) {
	if ctx == nil {
		return
	}
	ctx.displayMode = mode
}

// GetElementAt returns the element at the given screen coordinates.
func (ctx *GamepadImage) GetElementAt(x, y float32) int {
	if ctx == nil {
		return SDL_GAMEPAD_ELEMENT_INVALID
	}

	point := sdl.FPoint{X: x, Y: y}

	if ctx.showingFront {
		for i := 0; i < len(axisPositions); i++ {
			element := int(sdl.GAMEPAD_BUTTON_COUNT) + i

			if element == SDL_GAMEPAD_ELEMENT_AXIS_LEFT_TRIGGER ||
				element == SDL_GAMEPAD_ELEMENT_AXIS_RIGHT_TRIGGER {
				rect := sdl.FRect{
					W: ctx.AxisWidth,
					H: ctx.AxisHeight,
				}
				rect.X = ctx.x + float32(axisPositions[i].X) - rect.W/2
				rect.Y = ctx.y + float32(axisPositions[i].Y) - rect.H/2
				if PointInRectFloat(&point, &rect) {
					if element == SDL_GAMEPAD_ELEMENT_AXIS_LEFT_TRIGGER {
						return SDL_GAMEPAD_ELEMENT_AXIS_LEFT_TRIGGER
					}
					return SDL_GAMEPAD_ELEMENT_AXIS_RIGHT_TRIGGER
				}
			} else if element == SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_POSITIVE {
				rect := sdl.FRect{
					W: ctx.ButtonWidth * 2.0,
					H: ctx.ButtonHeight * 2.0,
				}
				rect.X = ctx.x + float32(buttonPositions[sdl.GAMEPAD_BUTTON_LEFT_STICK].X) - rect.W/2
				rect.Y = ctx.y + float32(buttonPositions[sdl.GAMEPAD_BUTTON_LEFT_STICK].Y) - rect.H/2
				if PointInRectFloat(&point, &rect) {
					thumbstickRadius := ctx.ButtonWidth * 0.1
					deltaX := x - (ctx.x + float32(buttonPositions[sdl.GAMEPAD_BUTTON_LEFT_STICK].X))
					deltaY := y - (ctx.y + float32(buttonPositions[sdl.GAMEPAD_BUTTON_LEFT_STICK].Y))
					deltaSq := deltaX*deltaX + deltaY*deltaY
					if deltaSq > thumbstickRadius*thumbstickRadius {
						angle := float32(math.Atan2(float64(deltaY), float64(deltaX))) + float32(math.Pi)
						if angle < float32(math.Pi)*0.25 {
							return SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE
						} else if angle < float32(math.Pi)*0.75 {
							return SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_NEGATIVE
						} else if angle < float32(math.Pi)*1.25 {
							return SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_POSITIVE
						} else if angle < float32(math.Pi)*1.75 {
							return SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_POSITIVE
						} else {
							return SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE
						}
					}
				}
			} else if element == SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_POSITIVE {
				rect := sdl.FRect{
					W: ctx.ButtonWidth * 2.0,
					H: ctx.ButtonHeight * 2.0,
				}
				rect.X = ctx.x + float32(buttonPositions[sdl.GAMEPAD_BUTTON_RIGHT_STICK].X) - rect.W/2
				rect.Y = ctx.y + float32(buttonPositions[sdl.GAMEPAD_BUTTON_RIGHT_STICK].Y) - rect.H/2
				if PointInRectFloat(&point, &rect) {
					thumbstickRadius := ctx.ButtonWidth * 0.1
					deltaX := x - (ctx.x + float32(buttonPositions[sdl.GAMEPAD_BUTTON_RIGHT_STICK].X))
					deltaY := y - (ctx.y + float32(buttonPositions[sdl.GAMEPAD_BUTTON_RIGHT_STICK].Y))
					deltaSq := deltaX*deltaX + deltaY*deltaY
					if deltaSq > thumbstickRadius*thumbstickRadius {
						angle := float32(math.Atan2(float64(deltaY), float64(deltaX))) + float32(math.Pi)
						if angle < float32(math.Pi)*0.25 {
							return SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_NEGATIVE
						} else if angle < float32(math.Pi)*0.75 {
							return SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_NEGATIVE
						} else if angle < float32(math.Pi)*1.25 {
							return SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_POSITIVE
						} else if angle < float32(math.Pi)*1.75 {
							return SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_POSITIVE
						} else {
							return SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_NEGATIVE
						}
					}
				}
			}
		}
	}

	for i := 0; i < len(buttonPositions); i++ {
		onFront := true
		if i >= int(sdl.GAMEPAD_BUTTON_RIGHT_PADDLE1) && i <= int(sdl.GAMEPAD_BUTTON_LEFT_PADDLE2) {
			onFront = false
		}
		if onFront == ctx.showingFront {
			rect := sdl.FRect{
				X: ctx.x + float32(buttonPositions[i].X) - ctx.ButtonWidth/2,
				Y: ctx.y + float32(buttonPositions[i].Y) - ctx.ButtonHeight/2,
				W: ctx.ButtonWidth,
				H: ctx.ButtonHeight,
			}
			if PointInRectFloat(&point, &rect) {
				return i
			}
		}
	}
	return SDL_GAMEPAD_ELEMENT_INVALID
}

// Clear resets all element states.
func (ctx *GamepadImage) Clear() {
	if ctx == nil {
		return
	}
	for i := range ctx.elements {
		ctx.elements[i] = false
	}
}

// SetElement sets the active state of an element.
func (ctx *GamepadImage) SetElement(element int, active bool) {
	if ctx == nil || element < 0 || element >= SDL_GAMEPAD_ELEMENT_MAX {
		return
	}
	ctx.elements[element] = active
}

// UpdateFromGamepad updates the image state from a live gamepad.
func (ctx *GamepadImage) UpdateFromGamepad(gamepad *sdl.Gamepad) {
	if ctx == nil || gamepad == nil {
		return
	}

	ctx.gamepadType = gamepad.Type()
	ctx.eastLabel = gamepad.ButtonLabel(sdl.GAMEPAD_BUTTON_EAST)
	mapping, err := gamepad.Mapping()
	if err == nil && mapping != "" {
		if strings.Contains(mapping, "SDL_GAMECONTROLLER_USE_BUTTON_LABELS") {
			ctx.gamepadType = sdl.GAMEPAD_TYPE_NINTENDO_SWITCH_PRO
		}
	}

	// Update button states
	for i := 0; i < int(sdl.GAMEPAD_BUTTON_TOUCHPAD); i++ {
		button := sdl.GamepadButton(i)
		ctx.SetElement(i, gamepad.Button(button))
	}

	// Update axis states
	const deadzone = 8000
	for i := 0; i < int(sdl.GAMEPAD_AXIS_COUNT); i++ {
		axis := sdl.GamepadAxis(i)
		value := gamepad.Axis(axis)
		switch axis {
		case sdl.GAMEPAD_AXIS_LEFTX:
			ctx.SetElement(SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE, value < -deadzone)
			ctx.SetElement(SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_POSITIVE, value > deadzone)
		case sdl.GAMEPAD_AXIS_RIGHTX:
			ctx.SetElement(SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_NEGATIVE, value < -deadzone)
			ctx.SetElement(SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_POSITIVE, value > deadzone)
		case sdl.GAMEPAD_AXIS_LEFTY:
			ctx.SetElement(SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_NEGATIVE, value < -deadzone)
			ctx.SetElement(SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_POSITIVE, value > deadzone)
		case sdl.GAMEPAD_AXIS_RIGHTY:
			ctx.SetElement(SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_NEGATIVE, value < -deadzone)
			ctx.SetElement(SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_POSITIVE, value > deadzone)
		case sdl.GAMEPAD_AXIS_LEFT_TRIGGER:
			ctx.SetElement(SDL_GAMEPAD_ELEMENT_AXIS_LEFT_TRIGGER, value > deadzone)
		case sdl.GAMEPAD_AXIS_RIGHT_TRIGGER:
			ctx.SetElement(SDL_GAMEPAD_ELEMENT_AXIS_RIGHT_TRIGGER, value > deadzone)
		}
	}

	ctx.connectionState, _ = gamepad.ConnectionState()
	state, pct := gamepad.PowerInfo()
	ctx.batteryState = state
	ctx.batteryPercent = int(pct)

	if gamepad.NumTouchpads() > 0 {
		numFingers := int(gamepad.NumTouchpadFingers(0))
		if numFingers != ctx.numFingers {
			ctx.fingers = make([]GamepadTouchpadFinger, numFingers)
			ctx.numFingers = numFingers
		}
		for i := 0; i < numFingers; i++ {
			var down bool
			var fx, fy, pressure float32
			gamepad.TouchpadFinger(0, int32(i), &down, &fx, &fy, &pressure)
			ctx.fingers[i] = GamepadTouchpadFinger{
				Down:     down,
				X:        fx,
				Y:        fy,
				Pressure: pressure,
			}
		}
		ctx.showingTouchpad = true
	} else {
		ctx.fingers = nil
		ctx.numFingers = 0
		ctx.showingTouchpad = false
	}
}

// Render draws the gamepad image.
func (ctx *GamepadImage) Render() {
	if ctx == nil {
		return
	}

	dst := sdl.FRect{X: ctx.x, Y: ctx.y, W: ctx.gamepadWidth, H: ctx.gamepadHeight}

	if ctx.showingFront {
		ctx.renderer.RenderTexture(ctx.frontTexture, nil, &dst)
	} else {
		ctx.renderer.RenderTexture(ctx.backTexture, nil, &dst)
	}

	// Draw active button elements
	for i := 0; i < len(buttonPositions); i++ {
		if ctx.elements[i] {
			onFront := true
			if i >= int(sdl.GAMEPAD_BUTTON_RIGHT_PADDLE1) && i <= int(sdl.GAMEPAD_BUTTON_LEFT_PADDLE2) {
				onFront = false
			}
			if onFront == ctx.showingFront {
				dst.W = ctx.ButtonWidth
				dst.H = ctx.ButtonHeight
				dst.X = ctx.x + float32(buttonPositions[i].X) - dst.W/2
				dst.Y = ctx.y + float32(buttonPositions[i].Y) - dst.H/2
				ctx.renderer.RenderTexture(ctx.buttonTexture, nil, &dst)
			}
		}
	}

	// Draw face buttons overlay
	if ctx.showingFront {
		dst.X = ctx.x + 363
		dst.Y = ctx.y + 118
		dst.W = ctx.faceWidth
		dst.H = ctx.faceHeight

		switch ctx.eastLabel {
		case sdl.GAMEPAD_BUTTON_LABEL_B:
			ctx.renderer.RenderTexture(ctx.faceABXYTexture, nil, &dst)
		case sdl.GAMEPAD_BUTTON_LABEL_X:
			ctx.renderer.RenderTexture(ctx.faceAXBYTexture, nil, &dst)
		case sdl.GAMEPAD_BUTTON_LABEL_A:
			ctx.renderer.RenderTexture(ctx.faceBAYXTexture, nil, &dst)
		case sdl.GAMEPAD_BUTTON_LABEL_CIRCLE:
			ctx.renderer.RenderTexture(ctx.faceSonyTexture, nil, &dst)
		}
	}

	// Draw active axis elements
	if ctx.showingFront {
		for i := 0; i < len(axisPositions); i++ {
			element := int(sdl.GAMEPAD_BUTTON_COUNT) + i
			if ctx.elements[element] {
				angle := axisPositions[i].Angle
				dst.W = ctx.AxisWidth
				dst.H = ctx.AxisHeight
				dst.X = ctx.x + float32(axisPositions[i].X) - dst.W/2
				dst.Y = ctx.y + float32(axisPositions[i].Y) - dst.H/2
				ctx.renderer.RenderTextureRotated(ctx.axisTexture, nil, &dst, angle, nil, sdl.FLIP_NONE)
			}
		}
	}

	// Draw connection indicator
	if ctx.displayMode == CONTROLLER_MODE_TESTING {
		dst.X = ctx.x + ctx.gamepadWidth - ctx.batteryWidth - 4 - ctx.connectionWidth
		dst.Y = ctx.y
		dst.W = ctx.connectionWidth
		dst.H = ctx.connectionHeight

		switch ctx.connectionState {
		case sdl.JOYSTICK_CONNECTION_WIRED:
			ctx.renderer.RenderTexture(ctx.connectionTexture[0], nil, &dst)
		case sdl.JOYSTICK_CONNECTION_WIRELESS:
			ctx.renderer.RenderTexture(ctx.connectionTexture[1], nil, &dst)
		}
	}

	// Draw battery indicator
	if ctx.displayMode == CONTROLLER_MODE_TESTING &&
		ctx.batteryState != sdl.POWERSTATE_NO_BATTERY &&
		ctx.batteryState != sdl.POWERSTATE_UNKNOWN {

		savedColor := saveColor(ctx.renderer)

		dst.X = ctx.x + ctx.gamepadWidth - ctx.batteryWidth
		dst.Y = ctx.y
		dst.W = ctx.batteryWidth
		dst.H = ctx.batteryHeight

		if ctx.batteryPercent > 40 {
			ctx.renderer.SetDrawColor(0x00, 0xD4, 0x50, 0xFF)
		} else if ctx.batteryPercent > 10 {
			ctx.renderer.SetDrawColor(0xFF, 0xC7, 0x00, 0xFF)
		} else {
			ctx.renderer.SetDrawColor(0xC8, 0x1D, 0x13, 0xFF)
		}

		fill := dst
		fill.X += 2
		fill.Y += 2
		fill.H -= 4
		fill.W = 25.0 * (float32(ctx.batteryPercent) / 100.0)
		ctx.renderer.RenderFillRect(&fill)
		restoreColor(ctx.renderer, savedColor)

		if ctx.batteryState == sdl.POWERSTATE_ON_BATTERY {
			ctx.renderer.RenderTexture(ctx.batteryTexture[0], nil, &dst)
		} else {
			ctx.renderer.RenderTexture(ctx.batteryTexture[1], nil, &dst)
		}
	}

	// Draw touchpad
	if ctx.displayMode == CONTROLLER_MODE_TESTING && ctx.showingTouchpad {
		dst.X = ctx.x + (ctx.gamepadWidth-ctx.touchpadWidth)/2
		dst.Y = ctx.y + ctx.gamepadHeight
		dst.W = ctx.touchpadWidth
		dst.H = ctx.touchpadHeight
		ctx.renderer.RenderTexture(ctx.touchpadTexture, nil, &dst)

		for i := 0; i < ctx.numFingers; i++ {
			finger := &ctx.fingers[i]
			if finger.Down {
				dst.X = ctx.x + (ctx.gamepadWidth-ctx.touchpadWidth)/2
				dst.X += touchpadArea.X + finger.X*touchpadArea.W
				dst.X -= ctx.ButtonWidth / 2
				dst.Y = ctx.y + ctx.gamepadHeight
				dst.Y += touchpadArea.Y + finger.Y*touchpadArea.H
				dst.Y -= ctx.ButtonHeight / 2
				dst.W = ctx.ButtonWidth
				dst.H = ctx.ButtonHeight
				ctx.buttonTexture.SetAlphaMod(uint8(finger.Pressure * 255))
				ctx.renderer.RenderTexture(ctx.buttonTexture, nil, &dst)
				ctx.buttonTexture.SetAlphaMod(255)
			}
		}
	}
}

// Destroy releases all image resources.
func (ctx *GamepadImage) Destroy() {
	if ctx == nil {
		return
	}
	if ctx.frontTexture != nil {
		ctx.frontTexture.Destroy()
	}
	if ctx.backTexture != nil {
		ctx.backTexture.Destroy()
	}
	if ctx.faceABXYTexture != nil {
		ctx.faceABXYTexture.Destroy()
	}
	if ctx.faceAXBYTexture != nil {
		ctx.faceAXBYTexture.Destroy()
	}
	if ctx.faceBAYXTexture != nil {
		ctx.faceBAYXTexture.Destroy()
	}
	if ctx.faceSonyTexture != nil {
		ctx.faceSonyTexture.Destroy()
	}
	for i := range ctx.connectionTexture {
		if ctx.connectionTexture[i] != nil {
			ctx.connectionTexture[i].Destroy()
		}
	}
	for i := range ctx.batteryTexture {
		if ctx.batteryTexture[i] != nil {
			ctx.batteryTexture[i].Destroy()
		}
	}
	if ctx.touchpadTexture != nil {
		ctx.touchpadTexture.Destroy()
	}
	if ctx.buttonTexture != nil {
		ctx.buttonTexture.Destroy()
	}
	if ctx.axisTexture != nil {
		ctx.axisTexture.Destroy()
	}
}
