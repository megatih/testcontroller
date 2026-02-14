package main

import (
	"fmt"

	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/megatih/testcontroller/gamepadutils"
)

func DrawGamepadWaiting(renderer *sdl.Renderer) {
	text := "Waiting for gamepad, press A to add a virtual controller"
	x := SCREEN_WIDTH/2 - (FONT_CHARACTER_SIZE*float32(len(text)))/2
	y := TITLE_HEIGHT/2 - FONT_CHARACTER_SIZE/2
	renderer.DebugText(x, y, text)
}

func DrawGamepadInfo(renderer *sdl.Renderer) {
	if controller == nil {
		return
	}

	if titleHighlighted {
		saved := gamepadutils.SaveColor(renderer)
		if titlePressed {
			renderer.SetDrawColor(gamepadutils.PRESSED_R, gamepadutils.PRESSED_G, gamepadutils.PRESSED_B, gamepadutils.PRESSED_A)
		} else {
			renderer.SetDrawColor(gamepadutils.HIGHLIGHT_R, gamepadutils.HIGHLIGHT_G, gamepadutils.HIGHLIGHT_B, gamepadutils.HIGHLIGHT_A)
		}
		renderer.RenderFillRect(&titleArea)
		gamepadutils.RestoreColor(renderer, saved)
	}

	if typeHighlighted {
		saved := gamepadutils.SaveColor(renderer)
		if typePressed {
			renderer.SetDrawColor(gamepadutils.PRESSED_R, gamepadutils.PRESSED_G, gamepadutils.PRESSED_B, gamepadutils.PRESSED_A)
		} else {
			renderer.SetDrawColor(gamepadutils.HIGHLIGHT_R, gamepadutils.HIGHLIGHT_G, gamepadutils.HIGHLIGHT_B, gamepadutils.HIGHLIGHT_A)
		}
		renderer.RenderFillRect(&typeArea)
		gamepadutils.RestoreColor(renderer, saved)
	}

	if controller.Joystick != nil {
		jid, _ := controller.Joystick.ID()
		text := fmt.Sprintf("(%d)", jid)
		x := SCREEN_WIDTH - (FONT_CHARACTER_SIZE*float32(len(text))) - 8.0
		y := float32(8.0)
		renderer.DebugText(x, y, text)
	}

	if controllerName != "" {
		x := titleArea.X + titleArea.W/2 - (FONT_CHARACTER_SIZE*float32(len(controllerName)))/2
		y := titleArea.Y + titleArea.H/2 - FONT_CHARACTER_SIZE/2
		renderer.DebugText(x, y, controllerName)
	}

	if controller.ID.IsJoystickVirtual() {
		text := "Click on the gamepad image below to generate input"
		x := SCREEN_WIDTH/2 - (FONT_CHARACTER_SIZE*float32(len(text)))/2
		y := TITLE_HEIGHT/2 - FONT_CHARACTER_SIZE/2 + FONT_LINE_HEIGHT + 2.0
		renderer.DebugText(x, y, text)
	}

	if controller.Gamepad != nil {
		typeStr := gamepadutils.GetGamepadTypeString(controller.Gamepad.Type())
		x := typeArea.X + typeArea.W/2 - (FONT_CHARACTER_SIZE*float32(len(typeStr)))/2
		y := typeArea.Y + typeArea.H/2 - FONT_CHARACTER_SIZE/2
		renderer.DebugText(x, y, typeStr)

		if displayMode == gamepadutils.CONTROLLER_MODE_TESTING {
			steamHandle := controller.Gamepad.SteamHandle()
			if steamHandle != 0 {
				text := fmt.Sprintf("Steam: 0x%.16x", steamHandle)
				y := SCREEN_HEIGHT - 2*(8.0+FONT_LINE_HEIGHT)
				x := SCREEN_WIDTH - 8.0 - (FONT_CHARACTER_SIZE * float32(len(text)))
				renderer.DebugText(x, y, text)
			}

			text := fmt.Sprintf("VID: 0x%.4x PID: 0x%.4x",
				controller.Joystick.Vendor(),
				controller.Joystick.Product())
			y2 := SCREEN_HEIGHT - 8.0 - FONT_LINE_HEIGHT
			x2 := SCREEN_WIDTH - 8.0 - (FONT_CHARACTER_SIZE * float32(len(text)))
			renderer.DebugText(x2, y2, text)

			serial := controller.Joystick.Serial()
			if serial != "" {
				text := fmt.Sprintf("Serial: %s", serial)
				x := SCREEN_WIDTH/2 - (FONT_CHARACTER_SIZE*float32(len(text)))/2
				y := SCREEN_HEIGHT - 8.0 - FONT_LINE_HEIGHT
				renderer.DebugText(x, y, text)
			}
		}
	}
}

func DrawBindingTips(renderer *sdl.Renderer) {
	imageArea := image.GetArea()
	buttonArea := doneMappingButton.GetArea()
	x := imageArea.X + imageArea.W/2
	y := imageArea.Y + imageArea.H
	y += (buttonArea.Y - y - FONT_CHARACTER_SIZE) / 2

	text := GetBindingInstruction()

	if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_INVALID {
		renderer.DebugText(x-(FONT_CHARACTER_SIZE*float32(len(text)))/2, y, text)
	} else {
		actionForward := sdl.GAMEPAD_BUTTON_SOUTH
		boundForward := gamepadutils.MappingHasElement(controller.Mapping, int(actionForward))
		actionBackward := sdl.GAMEPAD_BUTTON_EAST
		boundBackward := gamepadutils.MappingHasElement(controller.Mapping, int(actionBackward))
		actionDelete := sdl.GAMEPAD_BUTTON_WEST
		boundDelete := gamepadutils.MappingHasElement(controller.Mapping, int(actionDelete))

		y -= (FONT_CHARACTER_SIZE + BUTTON_MARGIN) / 2

		rect := sdl.FRect{
			W: 2.0 + (FONT_CHARACTER_SIZE * float32(len(text))) + 2.0,
			H: 2.0 + FONT_CHARACTER_SIZE + 2.0,
		}
		rect.X = x - rect.W/2
		rect.Y = y - 2.0

		saved := gamepadutils.SaveColor(renderer)
		renderer.SetDrawColor(gamepadutils.SELECTED_R, gamepadutils.SELECTED_G, gamepadutils.SELECTED_B, gamepadutils.SELECTED_A)
		renderer.RenderFillRect(&rect)
		gamepadutils.RestoreColor(renderer, saved)
		renderer.DebugText(x-(FONT_CHARACTER_SIZE*float32(len(text)))/2, y, text)

		y += FONT_CHARACTER_SIZE + BUTTON_MARGIN

		if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_NAME {
			text = "(press RETURN to complete)"
		} else if bindingElement == gamepadutils.SDL_GAMEPAD_ELEMENT_TYPE ||
			bindingElement == int(actionForward) ||
			bindingElement == int(actionBackward) {
			text = "(press ESC to cancel)"
		} else {
			gamepadTypeVal := image.GetType()
			if bindingFlow && boundForward && boundBackward {
				if bindingElement != int(actionDelete) && boundDelete {
					text = fmt.Sprintf("(press %s to skip, %s to go back, %s to delete, and ESC to cancel)",
						GetButtonLabel(gamepadTypeVal, actionForward),
						GetButtonLabel(gamepadTypeVal, actionBackward),
						GetButtonLabel(gamepadTypeVal, actionDelete))
				} else {
					text = fmt.Sprintf("(press %s to skip, %s to go back, SPACE to delete, and ESC to cancel)",
						GetButtonLabel(gamepadTypeVal, actionForward),
						GetButtonLabel(gamepadTypeVal, actionBackward))
				}
			} else {
				text = "(press SPACE to delete and ESC to cancel)"
			}
		}
		renderer.DebugText(x-(FONT_CHARACTER_SIZE*float32(len(text)))/2, y, text)
	}
}

func UpdateGamepadEffects() {
	if displayMode != gamepadutils.CONTROLLER_MODE_TESTING || controller == nil || controller.Gamepad == nil {
		return
	}

	// Update LED based on left thumbstick position
	x := controller.Gamepad.Axis(sdl.GAMEPAD_AXIS_LEFTX)
	y := controller.Gamepad.Axis(sdl.GAMEPAD_AXIS_LEFTY)

	if !setLED {
		setLED = x < -8000 || x > 8000 || y > 8000
	}
	if setLED {
		var r, g, b uint8
		if x < 0 {
			r = uint8(((int(^x)) * 255) / 32767)
		} else {
			b = uint8((int(x) * 255) / 32767)
		}
		if y > 0 {
			g = uint8((int(y) * 255) / 32767)
		}
		controller.Gamepad.SetLED(r, g, b)
	}

	if controller.TriggerEffect == 0 {
		// Update rumble based on trigger state
		left := controller.Gamepad.Axis(sdl.GAMEPAD_AXIS_LEFT_TRIGGER)
		right := controller.Gamepad.Axis(sdl.GAMEPAD_AXIS_RIGHT_TRIGGER)
		lowFreqRumble := convertAxisToRumble(left)
		highFreqRumble := convertAxisToRumble(right)
		controller.Gamepad.Rumble(lowFreqRumble, highFreqRumble, 250)

		// Update trigger rumble based on thumbstick state
		leftY := controller.Gamepad.Axis(sdl.GAMEPAD_AXIS_LEFTY)
		rightY := controller.Gamepad.Axis(sdl.GAMEPAD_AXIS_RIGHTY)
		leftRumble := convertAxisToRumble(^leftY)
		rightRumble := convertAxisToRumble(^rightY)
		controller.Gamepad.RumbleTriggers(leftRumble, rightRumble, 250)
	}
}

func ShowingFront() bool {
	if controller == nil || controller.Gamepad == nil {
		return true
	}

	showingFront := true
	for i := sdl.GAMEPAD_BUTTON_RIGHT_PADDLE1; i <= sdl.GAMEPAD_BUTTON_LEFT_PADDLE2; i++ {
		if controller.Gamepad.Button(i) || bindingElement == int(i) {
			showingFront = false
			break
		}
	}
	if (sdl.GetModState()&sdl.KMOD_SHIFT) != 0 && bindingElement != gamepadutils.SDL_GAMEPAD_ELEMENT_NAME {
		showingFront = false
	}
	return showingFront
}
