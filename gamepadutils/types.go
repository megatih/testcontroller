package gamepadutils

import "github.com/Zyko0/go-sdl3/sdl"

// ControllerDisplayMode determines what mode the display is in.
type ControllerDisplayMode int

const (
	CONTROLLER_MODE_TESTING ControllerDisplayMode = iota
	CONTROLLER_MODE_BINDING
)

// Gamepad element identifiers extending beyond SDL_GAMEPAD_BUTTON_COUNT.
const (
	SDL_GAMEPAD_ELEMENT_INVALID = -1

	// Buttons 0..GAMEPAD_BUTTON_COUNT-1 are SDL button enums used directly as int.

	SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE  = int(sdl.GAMEPAD_BUTTON_COUNT)
	SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_POSITIVE  = SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE + 1
	SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_NEGATIVE  = SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE + 2
	SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_POSITIVE  = SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE + 3
	SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_NEGATIVE = SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE + 4
	SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_POSITIVE = SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE + 5
	SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_NEGATIVE = SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE + 6
	SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_POSITIVE = SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE + 7
	SDL_GAMEPAD_ELEMENT_AXIS_LEFT_TRIGGER    = SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE + 8
	SDL_GAMEPAD_ELEMENT_AXIS_RIGHT_TRIGGER   = SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE + 9
	SDL_GAMEPAD_ELEMENT_AXIS_MAX             = SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE + 10

	SDL_GAMEPAD_ELEMENT_NAME = SDL_GAMEPAD_ELEMENT_AXIS_MAX
	SDL_GAMEPAD_ELEMENT_TYPE = SDL_GAMEPAD_ELEMENT_NAME + 1
	SDL_GAMEPAD_ELEMENT_MAX  = SDL_GAMEPAD_ELEMENT_TYPE + 1
)

// Color constants for UI highlighting.
const (
	HIGHLIGHT_R = 224
	HIGHLIGHT_G = 255
	HIGHLIGHT_B = 255
	HIGHLIGHT_A = 255

	PRESSED_R = 175
	PRESSED_G = 238
	PRESSED_B = 238
	PRESSED_A = 255

	SELECTED_R = 224
	SELECTED_G = 255
	SELECTED_B = 224
	SELECTED_A = 255
)

// Texture color mod values (no alpha).
var (
	HIGHLIGHT_TEXTURE_MOD = [3]uint8{224, 255, 255}
	PRESSED_TEXTURE_MOD   = [3]uint8{175, 238, 238}
)

// Gyro visualization colors.
var (
	GYRO_COLOR_RED    = [4]uint8{255, 0, 0, 255}
	GYRO_COLOR_GREEN  = [4]uint8{0, 255, 0, 255}
	GYRO_COLOR_BLUE   = [4]uint8{0, 0, 255, 255}
	GYRO_COLOR_ORANGE = [4]uint8{255, 128, 0, 255}
)

// Button layout constants.
const (
	BUTTON_PADDING       = 12.0
	MINIMUM_BUTTON_WIDTH = 96.0
)

// UTF-8 symbols.
const (
	DEGREE_UTF8  = "\xC2\xB0"
	SQUARED_UTF8 = "\xC2\xB2"
	MICRO_UTF8   = "\xC2\xB5"
)

// Accelerometer noise thresholds for gyro calibration.
const (
	ACCELEROMETER_NOISE_THRESHOLD = 1e-6
	ACCELEROMETER_MAX_NOISE_G     = 0.075
	ACCELEROMETER_MAX_NOISE_G_SQ  = ACCELEROMETER_MAX_NOISE_G * ACCELEROMETER_MAX_NOISE_G
)

// Quaternion represents a rotation quaternion used for IMU/gyro visualization.
type Quaternion struct {
	X, Y, Z, W float32
}

// EGyroCalibrationPhase represents the current gyro calibration state.
type EGyroCalibrationPhase int

const (
	GYRO_CALIBRATION_PHASE_OFF EGyroCalibrationPhase = iota
	GYRO_CALIBRATION_PHASE_NOISE_PROFILING
	GYRO_CALIBRATION_PHASE_DRIFT_PROFILING
	GYRO_CALIBRATION_PHASE_COMPLETE
)

// SDL_GAMEPAD_TYPE_UNSELECTED is used in the type display.
const SDL_GAMEPAD_TYPE_UNSELECTED = -1

// Font constants (matching SDL test font).
const (
	FONT_CHARACTER_SIZE = float32(sdl.DEBUG_TEXT_FONT_CHARACTER_SIZE)
	FONT_LINE_HEIGHT    = FONT_CHARACTER_SIZE + 2
)

// PointInRectFloat checks if a point is inside a floating-point rectangle.
func PointInRectFloat(p *sdl.FPoint, r *sdl.FRect) bool {
	return p.X >= r.X && p.X < r.X+r.W && p.Y >= r.Y && p.Y < r.Y+r.H
}

// GetGamepadTypeString returns the friendly name for a gamepad type.
func GetGamepadTypeString(t sdl.GamepadType) string {
	switch t {
	case sdl.GAMEPAD_TYPE_XBOX360:
		return "Xbox 360"
	case sdl.GAMEPAD_TYPE_XBOXONE:
		return "Xbox One"
	case sdl.GAMEPAD_TYPE_PS3:
		return "PS3"
	case sdl.GAMEPAD_TYPE_PS4:
		return "PS4"
	case sdl.GAMEPAD_TYPE_PS5:
		return "PS5"
	case sdl.GAMEPAD_TYPE_NINTENDO_SWITCH_PRO:
		return "Nintendo Switch"
	case sdl.GAMEPAD_TYPE_NINTENDO_SWITCH_JOYCON_LEFT:
		return "Joy-Con (L)"
	case sdl.GAMEPAD_TYPE_NINTENDO_SWITCH_JOYCON_RIGHT:
		return "Joy-Con (R)"
	case sdl.GAMEPAD_TYPE_NINTENDO_SWITCH_JOYCON_PAIR:
		return "Joy-Con Pair"
	default:
		return ""
	}
}
