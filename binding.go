package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/megatih/testcontroller/gamepadutils"
)

func SetCurrentBindingElement(element int, flow bool) {
	if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_NAME {
		RefreshControllerName()
	}

	if element == gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID {
		bindingFlowDirection = 0
		lastBindingElement = gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID
	} else {
		lastBindingElement = bindingElement
	}
	bindingElement = element
	bindingFlow = flow || (element == int(sdl.GAMEPAD_BUTTON_SOUTH))
	bindingAdvanceTime = 0

	if controller != nil {
		for i := 0; i < controller.NumAxes; i++ {
			controller.AxisStates[i].FarthestValue = controller.AxisStates[i].StartingValue
		}
	}

	gamepadElements.SetSelected(element)
}

func SetNextBindingElement() {
	if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID {
		return
	}

	for i := 0; i < len(bindingOrder); i++ {
		if bindingElement == bindingOrder[i] {
			bindingFlowDirection = 1
			SetCurrentBindingElement(bindingOrder[i+1], true)
			return
		}
	}
	SetCurrentBindingElement(gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID, false)
}

func SetPrevBindingElement() {
	if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID {
		return
	}

	for i := 1; i < len(bindingOrder); i++ {
		if bindingElement == bindingOrder[i] {
			bindingFlowDirection = -1
			SetCurrentBindingElement(bindingOrder[i-1], true)
			return
		}
	}
	SetCurrentBindingElement(gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID, false)
}

func StopBinding() {
	SetCurrentBindingElement(gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID, false)
}

type AxisInfo struct {
	Axis      int
	Direction int
}

func ParseAxisInfo(description string) (AxisInfo, bool) {
	if description == "" {
		return AxisInfo{}, false
	}

	info := AxisInfo{}
	s := description

	if s[0] == '-' {
		info.Direction = -1
		s = s[1:]
	} else if s[0] == '+' {
		info.Direction = 1
		s = s[1:]
	}

	if len(s) >= 2 && s[0] == 'a' && s[1] >= '0' && s[1] <= '9' {
		axis, err := strconv.Atoi(s[1:])
		if err == nil {
			info.Axis = axis
			return info, true
		}
	}
	return AxisInfo{}, false
}

func setAndFreeGamepadMapping(mapping string) {
	controller.ID.SetGamepadMapping(mapping)
}

func CommitBindingElement(binding string, force bool) {
	direction := 1
	ignoreBinding := false

	if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID {
		return
	}

	mapping := controller.Mapping

	// If the controller generates multiple events for a single element, pick the best one
	if !force && bindingAdvanceTime > 0 {
		current := gamepadutils.GetElementBinding(mapping, bindingElement)
		nativeButton := bindingElement < int(sdl.GAMEPAD_BUTTON_COUNT)
		nativeAxis := bindingElement >= int(sdl.GAMEPAD_BUTTON_COUNT) &&
			bindingElement <= gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_MAX
		nativeTrigger := bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_LEFT_TRIGGER ||
			bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_RIGHT_TRIGGER
		nativeDpad := bindingElement == int(sdl.GAMEPAD_BUTTON_DPAD_UP) ||
			bindingElement == int(sdl.GAMEPAD_BUTTON_DPAD_DOWN) ||
			bindingElement == int(sdl.GAMEPAD_BUTTON_DPAD_LEFT) ||
			bindingElement == int(sdl.GAMEPAD_BUTTON_DPAD_RIGHT)

		if nativeButton {
			currentButton := current != "" && current[0] == 'b'
			proposedButton := binding != "" && binding[0] == 'b'
			if currentButton && !proposedButton {
				ignoreBinding = true
			}
			if currentButton && proposedButton && len(current) > 1 && len(binding) > 1 && current[1] < binding[1] {
				ignoreBinding = true
			}
		}
		if nativeAxis {
			_, currentIsAxis := ParseAxisInfo(current)
			proposedAxisInfo, proposedIsAxis := ParseAxisInfo(binding)
			currentAxisInfo, _ := ParseAxisInfo(current)

			if currentIsAxis {
				ignoreBinding = true

				if nativeTrigger &&
					len(current) > 1 && len(binding) > 1 &&
					((current[0] == '-' && binding[0] == '+' && current[1:] == binding[1:]) ||
						(current[0] == '+' && binding[0] == '-' && current[1:] == binding[1:])) {
					binding = binding[1:]
					ignoreBinding = false
				}

				if proposedIsAxis && proposedAxisInfo.Axis < currentAxisInfo.Axis {
					ignoreBinding = false
				}
			}
		}
		if nativeDpad {
			currentHat := current != "" && current[0] == 'h'
			proposedHat := binding != "" && binding[0] == 'h'
			if currentHat && !proposedHat {
				ignoreBinding = true
			}
			if currentHat && proposedHat && len(current) > 1 && len(binding) > 1 && current[1] < binding[1] {
				ignoreBinding = true
			}
		}
	}

	if !ignoreBinding && bindingFlow && !force {
		existing := gamepadutils.GetElementForBinding(mapping, binding)
		if existing != gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID {
			actionForward := int(sdl.GAMEPAD_BUTTON_SOUTH)
			actionBackward := int(sdl.GAMEPAD_BUTTON_EAST)
			actionDelete := int(sdl.GAMEPAD_BUTTON_WEST)

			if bindingElement == actionForward {
				// Bind it!
			} else if bindingElement == actionBackward {
				if existing == actionForward {
					boundBackward := gamepadutils.MappingHasElement(controller.Mapping, actionBackward)
					if boundBackward {
						ignoreBinding = true
						SetNextBindingElement()
					} else {
						ignoreBinding = true
						SetPrevBindingElement()
					}
				} else if existing == actionBackward && bindingFlowDirection == -1 {
					ignoreBinding = true
					SetPrevBindingElement()
				} else {
					// Bind it!
				}
			} else if existing == actionForward {
				ignoreBinding = true
				SetNextBindingElement()
			} else if existing == actionBackward {
				ignoreBinding = true
				SetPrevBindingElement()
			} else if existing == bindingElement {
				ignoreBinding = true
				SetNextBindingElement()
			} else if existing == actionDelete {
				binding = ""
				direction = 1
				force = true
			} else if bindingElement != actionForward && bindingElement != actionBackward {
				// Clear the existing binding
			}
		}
	}

	if ignoreBinding {
		return
	}

	mapping = gamepadutils.ClearMappingBinding(mapping, binding)
	if binding != "" {
		mapping = gamepadutils.SetElementBinding(mapping, bindingElement, binding)
	} else {
		mapping = gamepadutils.SetElementBinding(mapping, bindingElement, "")
	}
	setAndFreeGamepadMapping(mapping)

	if force {
		if bindingFlow {
			if direction > 0 {
				SetNextBindingElement()
			} else if direction < 0 {
				SetPrevBindingElement()
			}
		} else {
			StopBinding()
		}
	} else {
		bindingAdvanceTime = sdl.Ticks() + 30
	}
}

func ClearBinding() {
	CommitBindingElement("", true)
}

func SetDisplayMode(mode gamepadutils.ControllerDisplayMode) {
	if mode == gamepadutils.CONTROLLER_MODE_BINDING {
		if controller != nil && controller.Mapping != "" {
			backupMapping = controller.Mapping
		}
		mappingController = controller.ID
		if gamepadutils.MappingHasBindings(backupMapping) {
			SetCurrentBindingElement(gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID, false)
		} else {
			SetCurrentBindingElement(int(sdl.GAMEPAD_BUTTON_SOUTH), true)
		}
	} else {
		backupMapping = ""
		mappingController = 0
		StopBinding()
	}

	displayMode = mode
	image.SetDisplayMode(mode)
	gamepadElements.SetDisplayMode(mode)

	buttonState, mx, my := sdl.GetMouseState()
	rx, ry, _ := screen.RenderCoordinatesFromWindow(mx, my)
	UpdateButtonHighlights(rx, ry, buttonState != 0)
}

func CancelMapping() {
	if backupMapping != "" {
		setAndFreeGamepadMapping(backupMapping)
		backupMapping = ""
	}
	SetDisplayMode(gamepadutils.CONTROLLER_MODE_TESTING)
}

func ClearMapping() {
	setAndFreeGamepadMapping("")
	SetCurrentBindingElement(gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID, false)
}

func CopyMapping() {
	if controller != nil && controller.Mapping != "" {
		sdl.SetClipboardText(controller.Mapping)
	}
}

func PasteMapping() {
	if controller != nil {
		mapping, _ := sdl.GetClipboardText()
		if gamepadutils.MappingHasBindings(mapping) {
			StopBinding()
			controller.ID.SetGamepadMapping(mapping)
			RefreshControllerName()
		}
	}
}

func CommitControllerName() {
	mapping := controller.Mapping
	mapping = gamepadutils.SetMappingName(mapping, controllerName)
	setAndFreeGamepadMapping(mapping)
}

func AddControllerNameText(text string) {
	controllerName += text
	CommitControllerName()
}

func BackspaceControllerName() {
	if len(controllerName) > 0 {
		controllerName = controllerName[:len(controllerName)-1]
	}
	CommitControllerName()
}

func ClearControllerName() {
	controllerName = ""
	CommitControllerName()
}

func CopyControllerName() {
	sdl.SetClipboardText(controllerName)
}

func PasteControllerName() {
	text, _ := sdl.GetClipboardText()
	controllerName = text
	CommitControllerName()
}

func CommitGamepadType(typ sdl.GamepadType) {
	mapping := controller.Mapping
	mapping = gamepadutils.SetMappingType(mapping, typ)
	setAndFreeGamepadMapping(mapping)
}

func GetBindingInstruction() string {
	switch bindingElement {
	case gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID:
		return "Select an element to bind from the list on the left"
	case int(sdl.GAMEPAD_BUTTON_SOUTH),
		int(sdl.GAMEPAD_BUTTON_EAST),
		int(sdl.GAMEPAD_BUTTON_WEST),
		int(sdl.GAMEPAD_BUTTON_NORTH):
		label := image.GetType().GamepadButtonLabelForType(sdl.GamepadButton(bindingElement))
		switch label {
		case sdl.GAMEPAD_BUTTON_LABEL_A:
			return "Press the A button"
		case sdl.GAMEPAD_BUTTON_LABEL_B:
			return "Press the B button"
		case sdl.GAMEPAD_BUTTON_LABEL_X:
			return "Press the X button"
		case sdl.GAMEPAD_BUTTON_LABEL_Y:
			return "Press the Y button"
		case sdl.GAMEPAD_BUTTON_LABEL_CROSS:
			return "Press the Cross (X) button"
		case sdl.GAMEPAD_BUTTON_LABEL_CIRCLE:
			return "Press the Circle button"
		case sdl.GAMEPAD_BUTTON_LABEL_SQUARE:
			return "Press the Square button"
		case sdl.GAMEPAD_BUTTON_LABEL_TRIANGLE:
			return "Press the Triangle button"
		default:
			return ""
		}
	case int(sdl.GAMEPAD_BUTTON_BACK):
		return "Press the left center button (Back/View/Share)"
	case int(sdl.GAMEPAD_BUTTON_GUIDE):
		return "Press the center button (Home/Guide)"
	case int(sdl.GAMEPAD_BUTTON_START):
		return "Press the right center button (Start/Menu/Options)"
	case int(sdl.GAMEPAD_BUTTON_LEFT_STICK):
		return "Press the left thumbstick button (LSB/L3)"
	case int(sdl.GAMEPAD_BUTTON_RIGHT_STICK):
		return "Press the right thumbstick button (RSB/R3)"
	case int(sdl.GAMEPAD_BUTTON_LEFT_SHOULDER):
		return "Press the left shoulder button (LB/L1)"
	case int(sdl.GAMEPAD_BUTTON_RIGHT_SHOULDER):
		return "Press the right shoulder button (RB/R1)"
	case int(sdl.GAMEPAD_BUTTON_DPAD_UP):
		return "Press the D-Pad up"
	case int(sdl.GAMEPAD_BUTTON_DPAD_DOWN):
		return "Press the D-Pad down"
	case int(sdl.GAMEPAD_BUTTON_DPAD_LEFT):
		return "Press the D-Pad left"
	case int(sdl.GAMEPAD_BUTTON_DPAD_RIGHT):
		return "Press the D-Pad right"
	case int(sdl.GAMEPAD_BUTTON_MISC1):
		return "Press the bottom center button (Share/Capture)"
	case int(sdl.GAMEPAD_BUTTON_RIGHT_PADDLE1):
		return "Press the upper paddle under your right hand"
	case int(sdl.GAMEPAD_BUTTON_LEFT_PADDLE1):
		return "Press the upper paddle under your left hand"
	case int(sdl.GAMEPAD_BUTTON_RIGHT_PADDLE2):
		return "Press the lower paddle under your right hand"
	case int(sdl.GAMEPAD_BUTTON_LEFT_PADDLE2):
		return "Press the lower paddle under your left hand"
	case int(sdl.GAMEPAD_BUTTON_TOUCHPAD):
		return "Press down on the touchpad"
	case int(sdl.GAMEPAD_BUTTON_MISC2),
		int(sdl.GAMEPAD_BUTTON_MISC3),
		int(sdl.GAMEPAD_BUTTON_MISC4),
		int(sdl.GAMEPAD_BUTTON_MISC5),
		int(sdl.GAMEPAD_BUTTON_MISC6):
		return "Press any additional button not already bound"
	case gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_NEGATIVE:
		return "Move the left thumbstick to the left"
	case gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_LEFTX_POSITIVE:
		return "Move the left thumbstick to the right"
	case gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_NEGATIVE:
		return "Move the left thumbstick up"
	case gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_LEFTY_POSITIVE:
		return "Move the left thumbstick down"
	case gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_NEGATIVE:
		return "Move the right thumbstick to the left"
	case gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_RIGHTX_POSITIVE:
		return "Move the right thumbstick to the right"
	case gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_NEGATIVE:
		return "Move the right thumbstick up"
	case gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_RIGHTY_POSITIVE:
		return "Move the right thumbstick down"
	case gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_LEFT_TRIGGER:
		return "Pull the left trigger (LT/L2)"
	case gamepadutils.SDL_GAMEPAD_ELEMENT_AXIS_RIGHT_TRIGGER:
		return "Pull the right trigger (RT/R2)"
	case gamepadutils.SDL_GAMEPAD_ELEMENT_NAME:
		return "Type the name of your controller"
	case gamepadutils.SDL_GAMEPAD_ELEMENT_TYPE:
		return "Select the type of your controller"
	default:
		return ""
	}
}

func GetButtonLabel(typ sdl.GamepadType, button sdl.GamepadButton) string {
	switch typ.GamepadButtonLabelForType(button) {
	case sdl.GAMEPAD_BUTTON_LABEL_A:
		return "A"
	case sdl.GAMEPAD_BUTTON_LABEL_B:
		return "B"
	case sdl.GAMEPAD_BUTTON_LABEL_X:
		return "X"
	case sdl.GAMEPAD_BUTTON_LABEL_Y:
		return "Y"
	case sdl.GAMEPAD_BUTTON_LABEL_CROSS:
		return "Cross (X)"
	case sdl.GAMEPAD_BUTTON_LABEL_CIRCLE:
		return "Circle"
	case sdl.GAMEPAD_BUTTON_LABEL_SQUARE:
		return "Square"
	case sdl.GAMEPAD_BUTTON_LABEL_TRIANGLE:
		return "Triangle"
	default:
		return "UNKNOWN"
	}
}

func ClearButtonHighlights() {
	titleHighlighted = false
	titlePressed = false
	typeHighlighted = false
	typePressed = false

	image.Clear()
	gamepadElements.SetHighlight(gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID, false)
	gamepadType.SetHighlight(gamepadutils.SDL_GAMEPAD_TYPE_UNSELECTED, false)
	gyroElements.ResetGyroButton.SetHighlight(false, false)
	gyroElements.CalibrateGyroButton.SetHighlight(false, false)
	setupMappingButton.SetHighlight(false, false)
	doneMappingButton.SetHighlight(false, false)
	cancelButton.SetHighlight(false, false)
	clearButton.SetHighlight(false, false)
	copyButton.SetHighlight(false, false)
	pasteButton.SetHighlight(false, false)
}

func UpdateButtonHighlights(x, y float32, buttonDown bool) {
	ClearButtonHighlights()
	gyroElements.ResetGyroButton.SetHighlight(gyroElements.ResetGyroButton.Contains(x, y), buttonDown)
	gyroElements.CalibrateGyroButton.SetHighlight(gyroElements.CalibrateGyroButton.Contains(x, y), buttonDown)

	if displayMode == gamepadutils.CONTROLLER_MODE_TESTING {
		setupMappingButton.SetHighlight(setupMappingButton.Contains(x, y), buttonDown)
	} else if displayMode == gamepadutils.CONTROLLER_MODE_BINDING {
		gamepadHighlightElement := gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID

		point := sdl.FPoint{X: x, Y: y}
		if gamepadutils.PointInRectFloat(&point, &titleArea) {
			titleHighlighted = true
			titlePressed = buttonDown
		}
		if gamepadutils.PointInRectFloat(&point, &typeArea) {
			typeHighlighted = true
			typePressed = buttonDown
		}

		if controller != nil && controller.Joystick != virtualJoystick {
			gamepadHighlightElement = image.GetElementAt(x, y)
		}
		if gamepadHighlightElement == gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID && controller != nil {
			gamepadHighlightElement = gamepadElements.GetElementAt(controller.Gamepad, x, y)
		}
		gamepadElements.SetHighlight(gamepadHighlightElement, buttonDown)

		if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_TYPE {
			gamepadHighlightType := gamepadType.GetTypeAt(x, y)
			gamepadType.SetHighlight(gamepadHighlightType, buttonDown)
		}

		if controller != nil {
			joystickHighlightElement := joystickElements.GetElementAt(controller.Joystick, x, y)
			joystickElements.SetHighlight(joystickHighlightElement, buttonDown)
		}

		doneMappingButton.SetHighlight(doneMappingButton.Contains(x, y), buttonDown)
		cancelButton.SetHighlight(cancelButton.Contains(x, y), buttonDown)
		clearButton.SetHighlight(clearButton.Contains(x, y), buttonDown)
		copyButton.SetHighlight(copyButton.Contains(x, y), buttonDown)
		pasteButton.SetHighlight(pasteButton.Contains(x, y), buttonDown)
	}
}

// Helper to strip any prefix +/- for axis comparison
func stripAxisPrefix(s string) string {
	if len(s) > 0 && (s[0] == '+' || s[0] == '-') {
		return s[1:]
	}
	return s
}

// Unused function kept for future reference
var _ = strings.Contains
var _ = fmt.Sprintf
