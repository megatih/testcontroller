package main

import (
	"fmt"
	"log"
	"math"
	"os"

	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/megatih/testcontroller/gamepadutils"
)

// Global state matching testcontroller.c static variables.
var (
	window *sdl.Window
	screen *sdl.Renderer
	done   bool
	setLED bool

	displayMode      = gamepadutils.CONTROLLER_MODE_TESTING
	image            *gamepadutils.GamepadImage
	gamepadElements  *gamepadutils.GamepadDisplay
	gyroElements     *gamepadutils.GyroDisplay
	gamepadType      *gamepadutils.GamepadTypeDisplay
	joystickElements *gamepadutils.JoystickDisplay

	setupMappingButton *gamepadutils.GamepadButton
	doneMappingButton  *gamepadutils.GamepadButton
	cancelButton       *gamepadutils.GamepadButton
	clearButton        *gamepadutils.GamepadButton
	copyButton         *gamepadutils.GamepadButton
	pasteButton        *gamepadutils.GamepadButton

	backupMapping string

	numControllers    int
	controllers       []Controller
	controller        *Controller
	mappingController sdl.JoystickID

	bindingElement       = gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID
	lastBindingElement   = gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID
	bindingFlow          bool
	bindingFlowDirection int
	bindingAdvanceTime   uint64

	titleArea        sdl.FRect
	titleHighlighted bool
	titlePressed     bool
	typeArea         sdl.FRect
	typeHighlighted  bool
	typePressed      bool

	controllerName string

	virtualJoystick       *sdl.Joystick
	virtualAxisActive     = sdl.GAMEPAD_AXIS_INVALID
	virtualAxisStartX     float32
	virtualAxisStartY     float32
	virtualButtonActive   = sdl.GAMEPAD_BUTTON_INVALID
	virtualTouchpadActive bool
	virtualTouchpadX      float32
	virtualTouchpadY      float32
)

// Binding order for guided mapping flow.
var bindingOrder = []int{
	// Standard sequence
	int(sdl.GAMEPAD_BUTTON_SOUTH),
	int(sdl.GAMEPAD_BUTTON_EAST),
	int(sdl.GAMEPAD_BUTTON_WEST),
	int(sdl.GAMEPAD_BUTTON_NORTH),
	int(sdl.GAMEPAD_BUTTON_DPAD_LEFT),
	int(sdl.GAMEPAD_BUTTON_DPAD_RIGHT),
	int(sdl.GAMEPAD_BUTTON_DPAD_UP),
	int(sdl.GAMEPAD_BUTTON_DPAD_DOWN),
	gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE,
	gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_POSITIVE,
	gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_NEGATIVE,
	gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_POSITIVE,
	int(sdl.GAMEPAD_BUTTON_LEFT_STICK),
	gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_NEGATIVE,
	gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_POSITIVE,
	gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_NEGATIVE,
	gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_POSITIVE,
	int(sdl.GAMEPAD_BUTTON_RIGHT_STICK),
	int(sdl.GAMEPAD_BUTTON_LEFT_SHOULDER),
	gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_LEFT_TRIGGER,
	int(sdl.GAMEPAD_BUTTON_RIGHT_SHOULDER),
	gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_RIGHT_TRIGGER,
	int(sdl.GAMEPAD_BUTTON_BACK),
	int(sdl.GAMEPAD_BUTTON_START),
	int(sdl.GAMEPAD_BUTTON_GUIDE),
	int(sdl.GAMEPAD_BUTTON_MISC1),
	gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID,

	// Paddle sequence
	int(sdl.GAMEPAD_BUTTON_RIGHT_PADDLE1),
	int(sdl.GAMEPAD_BUTTON_LEFT_PADDLE1),
	int(sdl.GAMEPAD_BUTTON_RIGHT_PADDLE2),
	int(sdl.GAMEPAD_BUTTON_LEFT_PADDLE2),
	gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID,
}

func getSensorName(sensor sdl.SensorType) string {
	switch sensor {
	case sdl.SENSOR_ACCEL:
		return "accelerometer"
	case sdl.SENSOR_GYRO:
		return "gyro"
	case sdl.SENSOR_ACCEL_L:
		return "accelerometer (L)"
	case sdl.SENSOR_GYRO_L:
		return "gyro (L)"
	case sdl.SENSOR_ACCEL_R:
		return "accelerometer (R)"
	case sdl.SENSOR_GYRO_R:
		return "gyro (R)"
	default:
		return "UNKNOWN"
	}
}

const (
	JOYSTICK_AXIS_MAX = 32767
	JOYSTICK_AXIS_MIN = -32768
	STANDARD_GRAVITY  = 9.80665
)

func convertAxisToRumble(axisval int16) uint16 {
	halfAxis := int16(math.Ceil(float64(JOYSTICK_AXIS_MAX) / 2.0))
	if axisval > halfAxis {
		return uint16(axisval-halfAxis) * 4
	}
	return 0
}

func standardizeAxisValue(nValue int) int {
	if nValue > JOYSTICK_AXIS_MAX/2 {
		return JOYSTICK_AXIS_MAX
	} else if nValue < JOYSTICK_AXIS_MIN/2 {
		return JOYSTICK_AXIS_MIN
	}
	return 0
}

func appInit() error {
	sdl.SetHint(sdl.HINT_JOYSTICK_HIDAPI, "1")
	sdl.SetHint(sdl.HINT_JOYSTICK_ENHANCED_REPORTS, "auto")
	sdl.SetHint(sdl.HINT_JOYSTICK_HIDAPI_STEAM, "1")
	sdl.SetHint(sdl.HINT_JOYSTICK_ROG_CHAKRAM, "1")
	sdl.SetHint(sdl.HINT_JOYSTICK_ALLOW_BACKGROUND_EVENTS, "1")
	sdl.SetHint(sdl.HINT_JOYSTICK_LINUX_DEADZONES, "1")

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_JOYSTICK | sdl.INIT_GAMEPAD); err != nil {
		return fmt.Errorf("couldn't initialize SDL: %w", err)
	}

	sdl.AddGamepadMappingsFromFile("assets/gamecontrollerdb.txt")

	contentScale, err := sdl.GetPrimaryDisplay().ContentScale()
	if err != nil || contentScale == 0 {
		contentScale = 1.0
	}
	screenWidth := int(math.Ceil(float64(SCREEN_WIDTH * contentScale)))
	screenHeight := int(math.Ceil(float64(SCREEN_HEIGHT * contentScale)))

	window, err = sdl.CreateWindow("SDL Controller Test", screenWidth, screenHeight, sdl.WINDOW_HIGH_PIXEL_DENSITY)
	if err != nil {
		return fmt.Errorf("couldn't create window: %w", err)
	}

	screen, err = window.CreateRenderer("")
	if err != nil {
		window.Destroy()
		return fmt.Errorf("couldn't create renderer: %w", err)
	}

	screen.SetDrawColor(0x00, 0x00, 0x00, sdl.ALPHA_OPAQUE)
	screen.Clear()
	screen.Present()

	screen.SetLogicalPresentation(int32(SCREEN_WIDTH), int32(SCREEN_HEIGHT), sdl.LOGICAL_PRESENTATION_LETTERBOX)

	titleArea = sdl.FRect{
		W: GAMEPAD_WIDTH,
		H: FONT_CHARACTER_SIZE + 2*BUTTON_MARGIN,
		X: PANEL_WIDTH + PANEL_SPACING,
	}
	titleArea.Y = TITLE_HEIGHT/2 - titleArea.H/2

	typeArea = sdl.FRect{
		W: PANEL_WIDTH - 2*BUTTON_MARGIN,
		H: FONT_CHARACTER_SIZE + 2*BUTTON_MARGIN,
		X: BUTTON_MARGIN,
	}
	typeArea.Y = TITLE_HEIGHT/2 - typeArea.H/2

	image = gamepadutils.CreateGamepadImage(screen)
	if image == nil {
		screen.Destroy()
		window.Destroy()
		return fmt.Errorf("couldn't create gamepad image")
	}
	image.SetPosition(PANEL_WIDTH+PANEL_SPACING, TITLE_HEIGHT)

	gamepadElements = gamepadutils.CreateGamepadDisplay(screen)
	area := sdl.FRect{X: 0, Y: TITLE_HEIGHT, W: PANEL_WIDTH, H: GAMEPAD_HEIGHT}
	gamepadElements.SetArea(&area)

	gyroElements = gamepadutils.CreateGyroDisplay(screen)
	vidReservedHeight := float32(24.0)
	area = sdl.FRect{
		W: SCREEN_WIDTH * 0.375,
		H: SCREEN_HEIGHT * 0.475,
	}
	area.X = SCREEN_WIDTH - area.W
	area.Y = SCREEN_HEIGHT - area.H - vidReservedHeight
	gyroElements.SetArea(&area)
	gamepadutils.InitCirclePoints3D()

	gamepadType = gamepadutils.CreateGamepadTypeDisplay(screen)
	area = sdl.FRect{X: 0, Y: TITLE_HEIGHT, W: PANEL_WIDTH, H: GAMEPAD_HEIGHT}
	gamepadType.SetArea(&area)

	joystickElements = gamepadutils.CreateJoystickDisplay(screen)
	area = sdl.FRect{
		X: PANEL_WIDTH + PANEL_SPACING + GAMEPAD_WIDTH + PANEL_SPACING,
		Y: TITLE_HEIGHT,
		W: PANEL_WIDTH,
		H: GAMEPAD_HEIGHT,
	}
	joystickElements.SetArea(&area)

	setupMappingButton = gamepadutils.CreateGamepadButton(screen, "Setup Mapping")
	area = sdl.FRect{
		W: max32(gamepadutils.MINIMUM_BUTTON_WIDTH, setupMappingButton.LabelWidth()+2*gamepadutils.BUTTON_PADDING),
		H: setupMappingButton.LabelHeight() + 2*gamepadutils.BUTTON_PADDING,
		X: BUTTON_MARGIN,
	}
	area.Y = SCREEN_HEIGHT - BUTTON_MARGIN - area.H
	setupMappingButton.SetArea(&area)

	cancelButton = gamepadutils.CreateGamepadButton(screen, "Cancel")
	area = sdl.FRect{
		W: max32(gamepadutils.MINIMUM_BUTTON_WIDTH, cancelButton.LabelWidth()+2*gamepadutils.BUTTON_PADDING),
		H: cancelButton.LabelHeight() + 2*gamepadutils.BUTTON_PADDING,
		X: BUTTON_MARGIN,
	}
	area.Y = SCREEN_HEIGHT - BUTTON_MARGIN - area.H
	cancelButton.SetArea(&area)

	clearButton = gamepadutils.CreateGamepadButton(screen, "Clear")
	area.X += area.W + gamepadutils.BUTTON_PADDING
	area.W = max32(gamepadutils.MINIMUM_BUTTON_WIDTH, clearButton.LabelWidth()+2*gamepadutils.BUTTON_PADDING)
	area.H = clearButton.LabelHeight() + 2*gamepadutils.BUTTON_PADDING
	area.Y = SCREEN_HEIGHT - BUTTON_MARGIN - area.H
	clearButton.SetArea(&area)

	copyButton = gamepadutils.CreateGamepadButton(screen, "Copy")
	area.X += area.W + gamepadutils.BUTTON_PADDING
	area.W = max32(gamepadutils.MINIMUM_BUTTON_WIDTH, copyButton.LabelWidth()+2*gamepadutils.BUTTON_PADDING)
	area.H = copyButton.LabelHeight() + 2*gamepadutils.BUTTON_PADDING
	area.Y = SCREEN_HEIGHT - BUTTON_MARGIN - area.H
	copyButton.SetArea(&area)

	pasteButton = gamepadutils.CreateGamepadButton(screen, "Paste")
	area.X += area.W + gamepadutils.BUTTON_PADDING
	area.W = max32(gamepadutils.MINIMUM_BUTTON_WIDTH, pasteButton.LabelWidth()+2*gamepadutils.BUTTON_PADDING)
	area.H = pasteButton.LabelHeight() + 2*gamepadutils.BUTTON_PADDING
	area.Y = SCREEN_HEIGHT - BUTTON_MARGIN - area.H
	pasteButton.SetArea(&area)

	doneMappingButton = gamepadutils.CreateGamepadButton(screen, "Done")
	area = sdl.FRect{
		W: max32(gamepadutils.MINIMUM_BUTTON_WIDTH, doneMappingButton.LabelWidth()+2*gamepadutils.BUTTON_PADDING),
		H: doneMappingButton.LabelHeight() + 2*gamepadutils.BUTTON_PADDING,
	}
	area.X = SCREEN_WIDTH/2 - area.W/2
	area.Y = SCREEN_HEIGHT - BUTTON_MARGIN - area.H
	doneMappingButton.SetArea(&area)

	// Do initial iteration to process initial gamepad list
	iterate()

	return nil
}

func handleEvent(event *sdl.Event) {
	screen.ConvertEventToRenderCoordinates(event)

	switch event.Type {
	case sdl.EVENT_JOYSTICK_ADDED:
		evt := event.JoyDeviceEvent()
		AddController(evt.Which, true)

	case sdl.EVENT_JOYSTICK_REMOVED:
		evt := event.JoyDeviceEvent()
		DelController(evt.Which)

	case sdl.EVENT_JOYSTICK_AXIS_MOTION:
		evt := event.JoyAxisEvent()
		if displayMode == gamepadutils.CONTROLLER_MODE_TESTING {
			if evt.Value <= -(int16(JOYSTICK_AXIS_MAX)/2) || evt.Value >= (int16(JOYSTICK_AXIS_MAX)/2) {
				SetController(evt.Which)
			}
		} else if displayMode == gamepadutils.CONTROLLER_MODE_BINDING &&
			controller != nil &&
			evt.Which == controller.ID &&
			int(evt.Axis) < controller.NumAxes &&
			bindingElement != gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID {

			const MAX_ALLOWED_JITTER = int(JOYSTICK_AXIS_MAX) / 80
			pAxisState := &controller.AxisStates[evt.Axis]
			nValue := int(evt.Value)

			if !pAxisState.Moving {
				nInitialValue, ok := controller.Joystick.AxisInitialState(int32(evt.Axis))
				pAxisState.Moving = ok
				pAxisState.LastValue = nValue
				pAxisState.StartingValue = int(nInitialValue)
				pAxisState.FarthestValue = int(nInitialValue)
			} else if abs(nValue-pAxisState.LastValue) <= MAX_ALLOWED_JITTER {
				break
			} else {
				pAxisState.LastValue = nValue
			}

			nCurrentDistance := abs(nValue - pAxisState.StartingValue)
			nFarthestDistance := abs(pAxisState.FarthestValue - pAxisState.StartingValue)
			if nCurrentDistance > nFarthestDistance {
				pAxisState.FarthestValue = nValue
				nFarthestDistance = abs(pAxisState.FarthestValue - pAxisState.StartingValue)
			}

			if nFarthestDistance >= 16000 && nCurrentDistance <= 10000 {
				axisMin := standardizeAxisValue(pAxisState.StartingValue)
				axisMax := standardizeAxisValue(pAxisState.FarthestValue)

				var binding string
				if axisMin == 0 && axisMax == int(JOYSTICK_AXIS_MIN) {
					binding = fmt.Sprintf("-a%d", evt.Axis)
				} else if axisMin == 0 && axisMax == int(JOYSTICK_AXIS_MAX) {
					binding = fmt.Sprintf("+a%d", evt.Axis)
				} else {
					binding = fmt.Sprintf("a%d", evt.Axis)
					if axisMin > axisMax {
						binding += "~"
					}
				}
				CommitBindingElement(binding, false)
			}
		}

	case sdl.EVENT_JOYSTICK_BUTTON_DOWN:
		if displayMode == gamepadutils.CONTROLLER_MODE_TESTING {
			evt := event.JoyButtonEvent()
			SetController(evt.Which)
		}

	case sdl.EVENT_JOYSTICK_BUTTON_UP:
		evt := event.JoyButtonEvent()
		if displayMode == gamepadutils.CONTROLLER_MODE_BINDING &&
			controller != nil &&
			evt.Which == controller.ID &&
			bindingElement != gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID {
			binding := fmt.Sprintf("b%d", evt.Button)
			CommitBindingElement(binding, false)
		}

	case sdl.EVENT_JOYSTICK_HAT_MOTION:
		evt := event.JoyHatEvent()
		if displayMode == gamepadutils.CONTROLLER_MODE_BINDING &&
			controller != nil &&
			evt.Which == controller.ID &&
			evt.Value != sdl.HAT_CENTERED &&
			bindingElement != gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID {
			binding := fmt.Sprintf("h%d.%d", evt.Hat, evt.Value)
			CommitBindingElement(binding, false)
		}

	case sdl.EVENT_GAMEPAD_ADDED:
		evt := event.GamepadDeviceEvent()
		HandleGamepadAdded(evt.Which, true)

	case sdl.EVENT_GAMEPAD_REMOVED:
		evt := event.GamepadDeviceEvent()
		HandleGamepadRemoved(evt.Which)

	case sdl.EVENT_GAMEPAD_REMAPPED:
		evt := event.GamepadDeviceEvent()
		HandleGamepadRemapped(evt.Which)

	case sdl.EVENT_GAMEPAD_STEAM_HANDLE_UPDATED:
		RefreshControllerName()

	case sdl.EVENT_GAMEPAD_SENSOR_UPDATE:
		HandleGamepadSensorEvent(event)

	case sdl.EVENT_GAMEPAD_BUTTON_DOWN, sdl.EVENT_GAMEPAD_BUTTON_UP:
		evt := event.GamepadButtonEvent()
		if displayMode == gamepadutils.CONTROLLER_MODE_TESTING {
			if event.Type == sdl.EVENT_GAMEPAD_BUTTON_DOWN {
				SetController(evt.Which)
			}
		}

		if displayMode == gamepadutils.CONTROLLER_MODE_TESTING {
			if event.Type == sdl.EVENT_GAMEPAD_BUTTON_DOWN &&
				controller != nil && controller.Gamepad != nil &&
				controller.Gamepad.Type() == sdl.GAMEPAD_TYPE_PS5 {
				if sdl.GamepadButton(evt.Button) == sdl.GAMEPAD_BUTTON_MISC1 {
					CyclePS5AudioRoute(controller)
				}
				if sdl.GamepadButton(evt.Button) == sdl.GAMEPAD_BUTTON_NORTH {
					CyclePS5TriggerEffect(controller)
				}
			}
		}

	case sdl.EVENT_MOUSE_BUTTON_DOWN:
		evt := event.MouseButtonEvent()
		if virtualJoystick != nil && controller != nil && controller.Joystick == virtualJoystick {
			VirtualGamepadMouseDown(evt.X, evt.Y)
		}
		UpdateButtonHighlights(evt.X, evt.Y, evt.Down)

	case sdl.EVENT_MOUSE_BUTTON_UP:
		evt := event.MouseButtonEvent()
		if virtualJoystick != nil && controller != nil && controller.Joystick == virtualJoystick {
			VirtualGamepadMouseUp(evt.X, evt.Y)
		}

		if displayMode == gamepadutils.CONTROLLER_MODE_TESTING {
			if controller != nil && gyroElements.ResetGyroButton.Contains(evt.X, evt.Y) {
				ResetGyroOrientation(controller.IMUState)
			} else if controller != nil && gyroElements.CalibrateGyroButton.Contains(evt.X, evt.Y) {
				BeginNoiseCalibrationPhase(controller.IMUState)
			} else if setupMappingButton.Contains(evt.X, evt.Y) {
				SetDisplayMode(gamepadutils.CONTROLLER_MODE_BINDING)
			}
		} else if displayMode == gamepadutils.CONTROLLER_MODE_BINDING {
			if doneMappingButton.Contains(evt.X, evt.Y) {
				if controller != nil && controller.Mapping != "" {
					log.Printf("Mapping complete:\n%s", controller.Mapping)
				}
				SetDisplayMode(gamepadutils.CONTROLLER_MODE_TESTING)
			} else if cancelButton.Contains(evt.X, evt.Y) {
				CancelMapping()
			} else if clearButton.Contains(evt.X, evt.Y) {
				ClearMapping()
			} else if controller != nil && controller.HasBindings &&
				copyButton.Contains(evt.X, evt.Y) {
				CopyMapping()
			} else if pasteButton.Contains(evt.X, evt.Y) {
				PasteMapping()
			} else if titlePressed {
				SetCurrentBindingElement(gamepadutils.SDL_GAMEPAD_ELEMENT_NAME, false)
			} else if typePressed {
				SetCurrentBindingElement(gamepadutils.SDL_GAMEPAD_ELEMENT_TYPE, false)
			} else if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_TYPE {
				typ := gamepadType.GetTypeAt(evt.X, evt.Y)
				if typ != gamepadutils.SDL_GAMEPAD_TYPE_UNSELECTED {
					CommitGamepadType(sdl.GamepadType(typ))
					StopBinding()
				}
			} else {
				gamepadElement := gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID

				if controller != nil && controller.Joystick != virtualJoystick {
					gamepadElement = image.GetElementAt(evt.X, evt.Y)
				}
				if gamepadElement == gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID && controller != nil {
					gamepadElement = gamepadElements.GetElementAt(controller.Gamepad, evt.X, evt.Y)
				}
				if gamepadElement != gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID {
					SetCurrentBindingElement(gamepadElement, true)
				}

				joystickElement := joystickElements.GetElementAt(controller.Joystick, evt.X, evt.Y)
				if joystickElement != "" {
					CommitBindingElement(joystickElement, true)
				}
			}
		}
		UpdateButtonHighlights(evt.X, evt.Y, evt.Down)

	case sdl.EVENT_MOUSE_MOTION:
		evt := event.MouseMotionEvent()
		if virtualJoystick != nil && controller != nil && controller.Joystick == virtualJoystick {
			VirtualGamepadMouseMotion(evt.X, evt.Y)
		}
		UpdateButtonHighlights(evt.X, evt.Y, evt.State != 0)

	case sdl.EVENT_KEY_DOWN:
		evt := event.KeyboardEvent()
		if displayMode == gamepadutils.CONTROLLER_MODE_TESTING {
			if evt.Key >= sdl.K_0 && evt.Key <= sdl.K_9 {
				if controller != nil && controller.Gamepad != nil {
					playerIndex := int32(evt.Key - sdl.K_0)
					controller.Gamepad.SetPlayerIndex(playerIndex)
				}
			} else if evt.Key == sdl.K_A {
				OpenVirtualGamepad()
			} else if evt.Key == sdl.K_D {
				CloseVirtualGamepad()
			} else if evt.Key == sdl.K_R && (evt.Mod&sdl.KMOD_CTRL) != 0 {
				sdl.ReloadGamepadMappings()
			} else if evt.Key == sdl.K_ESCAPE {
				done = true
			} else if evt.Key == sdl.K_SPACE {
				if controller != nil && controller.IMUState != nil {
					ResetGyroOrientation(controller.IMUState)
				}
			}
		} else if displayMode == gamepadutils.CONTROLLER_MODE_BINDING {
			if evt.Key == sdl.K_C && (evt.Mod&sdl.KMOD_CTRL) != 0 {
				if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_NAME {
					CopyControllerName()
				} else {
					CopyMapping()
				}
			} else if evt.Key == sdl.K_V && (evt.Mod&sdl.KMOD_CTRL) != 0 {
				if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_NAME {
					ClearControllerName()
					PasteControllerName()
				} else {
					PasteMapping()
				}
			} else if evt.Key == sdl.K_X && (evt.Mod&sdl.KMOD_CTRL) != 0 {
				if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_NAME {
					CopyControllerName()
					ClearControllerName()
				} else {
					CopyMapping()
					ClearMapping()
				}
			} else if evt.Key == sdl.K_SPACE {
				if bindingElement != gamepadutils.SDL_GAMEPAD_ELEMENT_NAME {
					ClearBinding()
				}
			} else if evt.Key == sdl.K_BACKSPACE {
				if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_NAME {
					BackspaceControllerName()
				}
			} else if evt.Key == sdl.K_RETURN {
				if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_NAME {
					StopBinding()
				}
			} else if evt.Key == sdl.K_ESCAPE {
				if bindingElement != gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID {
					StopBinding()
				} else {
					CancelMapping()
				}
			}
		}

	case sdl.EVENT_TEXT_INPUT:
		if displayMode == gamepadutils.CONTROLLER_MODE_BINDING {
			if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_NAME {
				evt := event.TextInputEvent()
				AddControllerNameText(evt.Text)
			}
		}

	case sdl.EVENT_QUIT:
		done = true
	}
}

func iterate() {
	// If we have a virtual controller, send virtual sensor readings
	if virtualJoystick != nil {
		accelData := []float32{0.0, STANDARD_GRAVITY, 0.0}
		gyroData := []float32{0.01, -0.01, 0.0}
		sensorTimestamp := sdl.TicksNS()
		virtualJoystick.SendVirtualSensorData(sdl.SENSOR_ACCEL, sensorTimestamp, accelData)
		virtualJoystick.SendVirtualSensorData(sdl.SENSOR_GYRO, sensorTimestamp, gyroData)
	}

	// Wait 30 ms for joystick events to stop coming in
	if bindingAdvanceTime > 0 && sdl.Ticks() > (bindingAdvanceTime+30) {
		if bindingFlow {
			SetNextBindingElement()
		} else {
			StopBinding()
		}
	}

	// Blank screen, set up for drawing this frame
	screen.SetDrawColor(0xFF, 0xFF, 0xFF, sdl.ALPHA_OPAQUE)
	screen.Clear()
	screen.SetDrawColor(0x10, 0x10, 0x10, sdl.ALPHA_OPAQUE)

	if controller != nil {
		image.SetShowingFront(ShowingFront())
		image.UpdateFromGamepad(controller.Gamepad)
		if displayMode == gamepadutils.CONTROLLER_MODE_BINDING &&
			bindingElement != gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID {
			image.SetElement(bindingElement, true)
		}
		image.Render()

		if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_TYPE {
			if controller.Gamepad != nil {
				gamepadType.SetRealType(controller.Gamepad.RealType())
			}
			gamepadType.Render()
		} else {
			gamepadElements.Render(controller.Gamepad)
		}
		joystickElements.Render(controller.Joystick)

		if displayMode == gamepadutils.CONTROLLER_MODE_TESTING {
			setupMappingButton.Render()
			gyroElements.Render(gamepadElements, controller.Gamepad)
		} else if displayMode == gamepadutils.CONTROLLER_MODE_BINDING {
			DrawBindingTips(screen)
			doneMappingButton.Render()
			cancelButton.Render()
			clearButton.Render()
			if controller.HasBindings {
				copyButton.Render()
			}
			pasteButton.Render()
		}

		DrawGamepadInfo(screen)
		UpdateGamepadEffects()
	} else {
		DrawGamepadWaiting(screen)
	}

	sdl.Delay(16)
	screen.Present()
}

func main() {
	if err := sdl.LoadLibrary(sdl.Path()); err != nil {
		log.Fatalf("Failed to load SDL3 library: %v", err)
	}
	defer sdl.CloseLibrary()

	// Parse command line args
	for _, arg := range os.Args[1:] {
		if arg == "--virtual" {
			defer func() {
				if virtualJoystick == nil {
					OpenVirtualGamepad()
				}
			}()
		}
	}

	if err := appInit(); err != nil {
		log.Fatalf("Initialization failed: %v", err)
	}
	defer appQuit()

	var event sdl.Event
	for !done {
		for sdl.PollEvent(&event) {
			handleEvent(&event)
			if done {
				break
			}
		}
		if !done {
			iterate()
		}
	}
}

func appQuit() {
	CloseVirtualGamepad()
	for numControllers > 0 {
		HandleGamepadRemoved(controllers[0].ID)
		DelController(controllers[0].ID)
	}
	controllers = nil

	if image != nil {
		image.Destroy()
	}
	if gamepadElements != nil {
		gamepadElements.Destroy()
	}
	if gyroElements != nil {
		gyroElements.Destroy()
	}
	if gamepadType != nil {
		gamepadType.Destroy()
	}
	if joystickElements != nil {
		joystickElements.Destroy()
	}
	if setupMappingButton != nil {
		setupMappingButton.Destroy()
	}
	if doneMappingButton != nil {
		doneMappingButton.Destroy()
	}
	if cancelButton != nil {
		cancelButton.Destroy()
	}
	if clearButton != nil {
		clearButton.Destroy()
	}
	if copyButton != nil {
		copyButton.Destroy()
	}
	if pasteButton != nil {
		pasteButton.Destroy()
	}

	if screen != nil {
		screen.Destroy()
	}
	if window != nil {
		window.Destroy()
	}
	sdl.Quit()
}

// Helper functions
func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
