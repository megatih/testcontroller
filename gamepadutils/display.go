package gamepadutils

import (
	"fmt"
	"math"
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"
)

// Gamepad button and axis display names.
var gamepadButtonNames = []string{
	"South",
	"East",
	"West",
	"North",
	"Back",
	"Guide",
	"Start",
	"Left Stick",
	"Right Stick",
	"Left Shoulder",
	"Right Shoulder",
	"DPAD Up",
	"DPAD Down",
	"DPAD Left",
	"DPAD Right",
	"Misc1",
	"Right Paddle 1",
	"Left Paddle 1",
	"Right Paddle 2",
	"Left Paddle 2",
	"Touchpad",
	"Misc2",
	"Misc3",
	"Misc4",
	"Misc5",
	"Misc6",
}

var gamepadAxisNames = []string{
	"LeftX",
	"LeftY",
	"RightX",
	"RightY",
	"Left Trigger",
	"Right Trigger",
}

// GamepadDisplay shows the list of gamepad buttons and axes with their states.
type GamepadDisplay struct {
	renderer      *sdl.Renderer
	buttonTexture *sdl.Texture
	arrowTexture  *sdl.Texture
	ButtonWidth   float32
	ButtonHeight  float32
	arrowWidth    float32
	arrowHeight   float32

	accelData               [3]float32
	gyroData                [3]float32
	GyroDriftCorrectionData [3]float32

	lastSensorUpdate uint64

	displayMode        ControllerDisplayMode
	elementHighlighted int
	elementPressed     bool
	elementSelected    int

	Area sdl.FRect
}

// CreateGamepadDisplay creates a new gamepad display widget.
func CreateGamepadDisplay(renderer *sdl.Renderer) *GamepadDisplay {
	ctx := &GamepadDisplay{
		renderer:           renderer,
		elementHighlighted: SDL_GAMEPAD_ELEMENT_INVALID,
		elementSelected:    SDL_GAMEPAD_ELEMENT_INVALID,
	}

	ctx.buttonTexture = CreateTextureFromPNG(renderer, gamepadButtonSmallPNG)
	if ctx.buttonTexture != nil {
		ctx.ButtonWidth, ctx.ButtonHeight, _ = ctx.buttonTexture.Size()
	}

	ctx.arrowTexture = CreateTextureFromPNG(renderer, gamepadAxisArrowPNG)
	if ctx.arrowTexture != nil {
		ctx.arrowWidth, ctx.arrowHeight, _ = ctx.arrowTexture.Size()
	}

	return ctx
}

// SetDisplayMode sets the display mode.
func (ctx *GamepadDisplay) SetDisplayMode(mode ControllerDisplayMode) {
	if ctx == nil {
		return
	}
	ctx.displayMode = mode
}

// SetArea sets the display area.
func (ctx *GamepadDisplay) SetArea(area *sdl.FRect) {
	if ctx == nil {
		return
	}
	ctx.Area = *area
}

// SetGyroDriftCorrection sets the gyro drift correction values.
func (ctx *GamepadDisplay) SetGyroDriftCorrection(data [3]float32) {
	if ctx == nil {
		return
	}
	ctx.GyroDriftCorrectionData = data
}

// SetHighlight sets the highlighted element.
func (ctx *GamepadDisplay) SetHighlight(element int, pressed bool) {
	if ctx == nil {
		return
	}
	ctx.elementHighlighted = element
	ctx.elementPressed = pressed
}

// SetSelected sets the selected element.
func (ctx *GamepadDisplay) SetSelected(element int) {
	if ctx == nil {
		return
	}
	ctx.elementSelected = element
}

// GetElementAt returns the element at the given coordinates.
func (ctx *GamepadDisplay) GetElementAt(gamepad *sdl.Gamepad, x, y float32) int {
	if ctx == nil {
		return SDL_GAMEPAD_ELEMENT_INVALID
	}

	const margin = 8.0
	center := ctx.Area.W / 2.0
	const arrowExtent = 48.0

	point := sdl.FPoint{X: x, Y: y}
	rect := sdl.FRect{
		X: ctx.Area.X + margin,
		Y: ctx.Area.Y + margin + FONT_CHARACTER_SIZE/2 - ctx.ButtonHeight/2,
		W: ctx.Area.W - margin*2,
		H: ctx.ButtonHeight,
	}

	for i := 0; i < int(sdl.GAMEPAD_BUTTON_COUNT); i++ {
		button := sdl.GamepadButton(i)
		if ctx.displayMode == CONTROLLER_MODE_TESTING && !gamepad.HasButton(button) {
			continue
		}
		if PointInRectFloat(&point, &rect) {
			return i
		}
		rect.Y += ctx.ButtonHeight + 2.0
	}

	for i := 0; i < int(sdl.GAMEPAD_AXIS_COUNT); i++ {
		axis := sdl.GamepadAxis(i)
		if ctx.displayMode == CONTROLLER_MODE_TESTING && !gamepad.HasAxis(axis) {
			continue
		}

		area := sdl.FRect{
			X: rect.X + center + 2.0,
			Y: rect.Y + FONT_CHARACTER_SIZE/2 - ctx.ButtonHeight/2,
			W: ctx.arrowWidth + arrowExtent,
			H: ctx.ButtonHeight,
		}

		if PointInRectFloat(&point, &area) {
			switch axis {
			case sdl.GAMEPAD_AXIS_LEFTX:
				return SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE
			case sdl.GAMEPAD_AXIS_LEFTY:
				return SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_NEGATIVE
			case sdl.GAMEPAD_AXIS_RIGHTX:
				return SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_NEGATIVE
			case sdl.GAMEPAD_AXIS_RIGHTY:
				return SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_NEGATIVE
			}
		}

		area.X += area.W
		if PointInRectFloat(&point, &area) {
			switch axis {
			case sdl.GAMEPAD_AXIS_LEFTX:
				return SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_POSITIVE
			case sdl.GAMEPAD_AXIS_LEFTY:
				return SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_POSITIVE
			case sdl.GAMEPAD_AXIS_RIGHTX:
				return SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_POSITIVE
			case sdl.GAMEPAD_AXIS_RIGHTY:
				return SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_POSITIVE
			case sdl.GAMEPAD_AXIS_LEFT_TRIGGER:
				return SDL_GAMEPAD_ELEMENT_AXIS_LEFT_TRIGGER
			case sdl.GAMEPAD_AXIS_RIGHT_TRIGGER:
				return SDL_GAMEPAD_ELEMENT_AXIS_RIGHT_TRIGGER
			}
		}

		rect.Y += ctx.ButtonHeight + 2.0
	}
	return SDL_GAMEPAD_ELEMENT_INVALID
}

// renderElementHighlight draws a highlight rectangle for a specific element.
func (ctx *GamepadDisplay) renderElementHighlight(element int, area *sdl.FRect) {
	if element == ctx.elementHighlighted || element == ctx.elementSelected {
		savedColor := saveColor(ctx.renderer)

		if element == ctx.elementHighlighted {
			if ctx.elementPressed {
				ctx.renderer.SetDrawColor(PRESSED_R, PRESSED_G, PRESSED_B, PRESSED_A)
			} else {
				ctx.renderer.SetDrawColor(HIGHLIGHT_R, HIGHLIGHT_G, HIGHLIGHT_B, HIGHLIGHT_A)
			}
		} else {
			ctx.renderer.SetDrawColor(SELECTED_R, SELECTED_G, SELECTED_B, SELECTED_A)
		}
		ctx.renderer.RenderFillRect(area)
		restoreColor(ctx.renderer, savedColor)
	}
}

// getBindingString finds a binding value in a mapping string for a given label.
func getBindingString(label, mapping string) (string, bool) {
	var results []string
	idx := strings.Index(mapping, label)
	for idx >= 0 {
		value := mapping[idx+len(label):]
		end := strings.Index(value, ",")
		if end >= 0 {
			value = value[:end]
		}
		results = append(results, value)
		rest := mapping[idx+len(label):]
		nextIdx := strings.Index(rest, label)
		if nextIdx < 0 {
			break
		}
		idx = idx + len(label) + nextIdx
	}
	if len(results) == 0 {
		return "", false
	}
	return strings.Join(results, ","), true
}

// getButtonBindingString gets the binding string for a button.
func getButtonBindingString(button sdl.GamepadButton, mapping string) (string, bool) {
	label := fmt.Sprintf(",%s:", button.GamepadStringForButton())
	if text, ok := getBindingString(label, mapping); ok {
		return text, true
	}

	baxyMapping := strings.Contains(mapping, ",hint:SDL_GAMECONTROLLER_USE_BUTTON_LABELS:=1")
	switch button {
	case sdl.GAMEPAD_BUTTON_SOUTH:
		if baxyMapping {
			return getBindingString(",b:", mapping)
		}
		return getBindingString(",a:", mapping)
	case sdl.GAMEPAD_BUTTON_EAST:
		if baxyMapping {
			return getBindingString(",a:", mapping)
		}
		return getBindingString(",b:", mapping)
	case sdl.GAMEPAD_BUTTON_WEST:
		if baxyMapping {
			return getBindingString(",y:", mapping)
		}
		return getBindingString(",x:", mapping)
	case sdl.GAMEPAD_BUTTON_NORTH:
		if baxyMapping {
			return getBindingString(",x:", mapping)
		}
		return getBindingString(",y:", mapping)
	}
	return "", false
}

// getAxisBindingString gets the binding string for an axis direction.
func getAxisBindingString(axis sdl.GamepadAxis, direction int, mapping string) (string, bool) {
	// Check explicit half-axis
	var label string
	if direction < 0 {
		label = fmt.Sprintf(",-%s:", axis.GamepadStringForAxis())
	} else {
		label = fmt.Sprintf(",+%s:", axis.GamepadStringForAxis())
	}
	if text, ok := getBindingString(label, mapping); ok {
		return text, true
	}

	// Get whole axis binding
	label = fmt.Sprintf(",%s:", axis.GamepadStringForAxis())
	text, ok := getBindingString(label, mapping)
	if !ok {
		return "", false
	}
	if axis != sdl.GAMEPAD_AXIS_LEFT_TRIGGER && axis != sdl.GAMEPAD_AXIS_RIGHT_TRIGGER {
		if len(text) > 0 && text[0] == 'a' {
			// Split the axis
			if len(text) > 0 && text[len(text)-1] == '~' {
				direction *= -1
				text = text[:len(text)-1]
			}
			if direction > 0 {
				text = "+" + text
			} else {
				text = "-" + text
			}
		}
	}
	return text, true
}

// Render draws the gamepad display.
func (ctx *GamepadDisplay) Render(gamepad *sdl.Gamepad) {
	if ctx == nil {
		return
	}

	savedColor := saveColor(ctx.renderer)

	mapping, _ := gamepad.Mapping()

	const margin = 8.0
	center := ctx.Area.W / 2.0
	const arrowExtent = 48.0

	x := ctx.Area.X + margin
	y := ctx.Area.Y + margin

	for i := 0; i < int(sdl.GAMEPAD_BUTTON_COUNT); i++ {
		button := sdl.GamepadButton(i)
		if ctx.displayMode == CONTROLLER_MODE_TESTING && !gamepad.HasButton(button) {
			continue
		}

		highlight := sdl.FRect{
			X: x,
			Y: y + FONT_CHARACTER_SIZE/2 - ctx.ButtonHeight/2,
			W: ctx.Area.W - margin*2,
			H: ctx.ButtonHeight,
		}
		ctx.renderElementHighlight(i, &highlight)

		text := gamepadButtonNames[i] + ":"
		ctx.renderer.DebugText(x+center-float32(len(text))*FONT_CHARACTER_SIZE, y, text)

		if gamepad.Button(button) {
			ctx.buttonTexture.SetColorMod(10, 255, 21)
		} else {
			ctx.buttonTexture.SetColorMod(255, 255, 255)
		}

		dst := sdl.FRect{
			X: x + center + 2.0,
			Y: y + FONT_CHARACTER_SIZE/2 - ctx.ButtonHeight/2,
			W: ctx.ButtonWidth,
			H: ctx.ButtonHeight,
		}
		ctx.renderer.RenderTexture(ctx.buttonTexture, nil, &dst)

		if ctx.displayMode == CONTROLLER_MODE_BINDING {
			if binding, ok := getButtonBindingString(button, mapping); ok {
				ctx.renderer.DebugText(dst.X+dst.W+2*margin, y, binding)
			}
		}

		y += ctx.ButtonHeight + 2.0
	}

	for i := 0; i < int(sdl.GAMEPAD_AXIS_COUNT); i++ {
		axis := sdl.GamepadAxis(i)
		hasNegative := axis != sdl.GAMEPAD_AXIS_LEFT_TRIGGER && axis != sdl.GAMEPAD_AXIS_RIGHT_TRIGGER

		if ctx.displayMode == CONTROLLER_MODE_TESTING && !gamepad.HasAxis(axis) {
			continue
		}

		value := gamepad.Axis(axis)

		text := gamepadAxisNames[i] + ":"
		ctx.renderer.DebugText(x+center-float32(len(text))*FONT_CHARACTER_SIZE, y, text)

		highlight := sdl.FRect{
			X: x + center + 2.0,
			Y: y + FONT_CHARACTER_SIZE/2 - ctx.ButtonHeight/2,
			W: ctx.arrowWidth + arrowExtent,
			H: ctx.ButtonHeight,
		}

		switch axis {
		case sdl.GAMEPAD_AXIS_LEFTX:
			ctx.renderElementHighlight(SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE, &highlight)
		case sdl.GAMEPAD_AXIS_LEFTY:
			ctx.renderElementHighlight(SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_NEGATIVE, &highlight)
		case sdl.GAMEPAD_AXIS_RIGHTX:
			ctx.renderElementHighlight(SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_NEGATIVE, &highlight)
		case sdl.GAMEPAD_AXIS_RIGHTY:
			ctx.renderElementHighlight(SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_NEGATIVE, &highlight)
		}

		highlight.X += highlight.W

		switch axis {
		case sdl.GAMEPAD_AXIS_LEFTX:
			ctx.renderElementHighlight(SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_POSITIVE, &highlight)
		case sdl.GAMEPAD_AXIS_LEFTY:
			ctx.renderElementHighlight(SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_POSITIVE, &highlight)
		case sdl.GAMEPAD_AXIS_RIGHTX:
			ctx.renderElementHighlight(SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_POSITIVE, &highlight)
		case sdl.GAMEPAD_AXIS_RIGHTY:
			ctx.renderElementHighlight(SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_POSITIVE, &highlight)
		case sdl.GAMEPAD_AXIS_LEFT_TRIGGER:
			ctx.renderElementHighlight(SDL_GAMEPAD_ELEMENT_AXIS_LEFT_TRIGGER, &highlight)
		case sdl.GAMEPAD_AXIS_RIGHT_TRIGGER:
			ctx.renderElementHighlight(SDL_GAMEPAD_ELEMENT_AXIS_RIGHT_TRIGGER, &highlight)
		}

		dst := sdl.FRect{
			X: x + center + 2.0,
			Y: y + FONT_CHARACTER_SIZE/2 - ctx.arrowHeight/2,
			W: ctx.arrowWidth,
			H: ctx.arrowHeight,
		}

		if hasNegative {
			if value == math.MinInt16 {
				ctx.arrowTexture.SetColorMod(10, 255, 21)
			} else {
				ctx.arrowTexture.SetColorMod(255, 255, 255)
			}
			ctx.renderer.RenderTextureRotated(ctx.arrowTexture, nil, &dst, 0.0, nil, sdl.FLIP_HORIZONTAL)
		}

		dst.X += ctx.arrowWidth

		ctx.renderer.SetDrawColor(200, 200, 200, 255)
		rect := sdl.FRect{
			X: dst.X + arrowExtent - 2.0,
			Y: dst.Y,
			W: 4.0,
			H: ctx.arrowHeight,
		}
		ctx.renderer.RenderFillRect(&rect)
		restoreColor(ctx.renderer, savedColor)

		if value < 0 {
			ctx.renderer.SetDrawColor(8, 200, 16, 255)
			rect.W = (float32(value) / float32(math.MinInt16)) * arrowExtent
			rect.X = dst.X + arrowExtent - rect.W
			rect.Y = dst.Y + ctx.arrowHeight*0.25
			rect.H = ctx.arrowHeight / 2.0
			ctx.renderer.RenderFillRect(&rect)
		}

		if ctx.displayMode == CONTROLLER_MODE_BINDING && hasNegative {
			if binding, ok := getAxisBindingString(axis, -1, mapping); ok {
				restoreColor(ctx.renderer, savedColor)
				textX := dst.X + arrowExtent/2 - (FONT_CHARACTER_SIZE*float32(len(binding)))/2
				ctx.renderer.DebugText(textX, y, binding)
			}
		}

		dst.X += arrowExtent

		if value > 0 {
			ctx.renderer.SetDrawColor(8, 200, 16, 255)
			rect.W = (float32(value) / float32(math.MaxInt16)) * arrowExtent
			rect.X = dst.X
			rect.Y = dst.Y + ctx.arrowHeight*0.25
			rect.H = ctx.arrowHeight / 2.0
			ctx.renderer.RenderFillRect(&rect)
		}

		if ctx.displayMode == CONTROLLER_MODE_BINDING {
			if binding, ok := getAxisBindingString(axis, 1, mapping); ok {
				restoreColor(ctx.renderer, savedColor)
				textX := dst.X + arrowExtent/2 - (FONT_CHARACTER_SIZE*float32(len(binding)))/2
				ctx.renderer.DebugText(textX, y, binding)
			}
		}

		dst.X += arrowExtent

		if value == math.MaxInt16 {
			ctx.arrowTexture.SetColorMod(10, 255, 21)
		} else {
			ctx.arrowTexture.SetColorMod(255, 255, 255)
		}
		ctx.renderer.RenderTexture(ctx.arrowTexture, nil, &dst)

		restoreColor(ctx.renderer, savedColor)

		y += ctx.ButtonHeight + 2.0
	}

	if ctx.displayMode == CONTROLLER_MODE_TESTING {
		if gamepad.NumTouchpads() > 0 {
			numFingers := int(gamepad.NumTouchpadFingers(0))
			for i := 0; i < numFingers; i++ {
				var down bool
				var fingerX, fingerY, fingerPressure float32
				if !gamepad.TouchpadFinger(0, int32(i), &down, &fingerX, &fingerY, &fingerPressure) {
					continue
				}
				_ = fingerPressure

				text := fmt.Sprintf("Touch finger %d:", i)
				ctx.renderer.DebugText(x+center-float32(len(text))*FONT_CHARACTER_SIZE, y, text)

				if down {
					ctx.buttonTexture.SetColorMod(10, 255, 21)
				} else {
					ctx.buttonTexture.SetColorMod(255, 255, 255)
				}

				dst := sdl.FRect{
					X: x + center + 2.0,
					Y: y + FONT_CHARACTER_SIZE/2 - ctx.ButtonHeight/2,
					W: ctx.ButtonWidth,
					H: ctx.ButtonHeight,
				}
				ctx.renderer.RenderTexture(ctx.buttonTexture, nil, &dst)

				if down {
					posText := fmt.Sprintf("(%.2f,%.2f)", fingerX, fingerY)
					ctx.renderer.DebugText(x+center+ctx.ButtonWidth+4.0, y, posText)
				}

				y += ctx.ButtonHeight + 2.0
			}
		}

		hasAccel := gamepad.HasSensor(sdl.SENSOR_ACCEL)
		hasGyro := gamepad.HasSensor(sdl.SENSOR_GYRO)

		if hasAccel || hasGyro {
			const sensorUpdateIntervalMS = 100
			now := sdl.Ticks()

			if now >= ctx.lastSensorUpdate+sensorUpdateIntervalMS {
				if hasAccel {
					gamepad.SensorData(sdl.SENSOR_ACCEL, &ctx.accelData[0], 3)
				}
				if hasGyro {
					gamepad.SensorData(sdl.SENSOR_GYRO, &ctx.gyroData[0], 3)
				}
				ctx.lastSensorUpdate = now
			}

			if hasAccel {
				text := "Accelerometer:"
				ctx.renderer.DebugText(x+center-float32(len(text))*FONT_CHARACTER_SIZE, y, text)
				text = fmt.Sprintf("[%.2f,%.2f,%.2f]m/s%s", ctx.accelData[0], ctx.accelData[1], ctx.accelData[2], SQUARED_UTF8)
				ctx.renderer.DebugText(x+center+2.0, y, text)
				y += ctx.ButtonHeight + 2.0
			}

			if hasGyro {
				radToDeg := float32(180.0 / math.Pi)
				text := "Gyro:"
				ctx.renderer.DebugText(x+center-float32(len(text))*FONT_CHARACTER_SIZE, y, text)
				text = fmt.Sprintf("[%.2f,%.2f,%.2f]%s/s", ctx.gyroData[0]*radToDeg, ctx.gyroData[1]*radToDeg, ctx.gyroData[2]*radToDeg, DEGREE_UTF8)
				ctx.renderer.DebugText(x+center+2.0, y, text)

				if ctx.GyroDriftCorrectionData[0] != 0 || ctx.GyroDriftCorrectionData[1] != 0 || ctx.GyroDriftCorrectionData[2] != 0 {
					y += ctx.ButtonHeight + 2.0
					text = "Gyro Drift:"
					ctx.renderer.DebugText(x+center-float32(len(text))*FONT_CHARACTER_SIZE, y, text)
					text = fmt.Sprintf("[%.2f,%.2f,%.2f]%s/s", ctx.GyroDriftCorrectionData[0]*radToDeg, ctx.GyroDriftCorrectionData[1]*radToDeg, ctx.GyroDriftCorrectionData[2]*radToDeg, DEGREE_UTF8)
					ctx.renderer.DebugText(x+center+2.0, y, text)
				}
			}
		}
	}
}

// Destroy releases the display resources.
func (ctx *GamepadDisplay) Destroy() {
	if ctx == nil {
		return
	}
	if ctx.buttonTexture != nil {
		ctx.buttonTexture.Destroy()
	}
	if ctx.arrowTexture != nil {
		ctx.arrowTexture.Destroy()
	}
}
