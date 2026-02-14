package main

import (
	"log"
	"math"

	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/megatih/testcontroller/gamepadutils"
)

func OpenVirtualGamepad() {
	if virtualJoystick != nil {
		return
	}

	desc := sdl.VirtualJoystickDesc{
		Type:     sdl.JOYSTICK_TYPE_GAMEPAD,
		Naxes:    uint16(sdl.GAMEPAD_AXIS_COUNT),
		Nbuttons: uint16(sdl.GAMEPAD_BUTTON_COUNT),
		Touchpads: []sdl.VirtualJoystickTouchpadDesc{
			{Nfingers: 1},
		},
		Sensors: []sdl.VirtualJoystickSensorDesc{
			{Type: sdl.SENSOR_ACCEL},
			{Type: sdl.SENSOR_GYRO},
		},
	}

	virtualID, err := sdl.AttachVirtualJoystick(&desc)
	if err != nil {
		log.Printf("Couldn't attach virtual device: %v", err)
		return
	}

	joystick, err := virtualID.OpenJoystick()
	if err != nil {
		log.Printf("Couldn't open virtual device: %v", err)
		return
	}
	virtualJoystick = joystick
}

func CloseVirtualGamepad() {
	joysticks, err := sdl.GetJoysticks()
	if err == nil {
		for _, id := range joysticks {
			if id.IsJoystickVirtual() {
				id.DetachVirtualJoystick()
			}
		}
	}

	if virtualJoystick != nil {
		virtualJoystick.Close()
		virtualJoystick = nil
	}
}

func VirtualGamepadMouseMotion(x, y float32) {
	if virtualButtonActive != sdl.GAMEPAD_BUTTON_INVALID {
		if virtualAxisActive != sdl.GAMEPAD_AXIS_INVALID {
			const MOVING_DISTANCE = 2.0
			if float32(math.Abs(float64(x-virtualAxisStartX))) >= MOVING_DISTANCE ||
				float32(math.Abs(float64(y-virtualAxisStartY))) >= MOVING_DISTANCE {
				virtualJoystick.SetVirtualButton(int32(virtualButtonActive), false)
				virtualButtonActive = sdl.GAMEPAD_BUTTON_INVALID
			}
		}
	}

	if virtualAxisActive != sdl.GAMEPAD_AXIS_INVALID {
		if virtualAxisActive == sdl.GAMEPAD_AXIS_LEFT_TRIGGER ||
			virtualAxisActive == sdl.GAMEPAD_AXIS_RIGHT_TRIGGER {
			axisRange := int(JOYSTICK_AXIS_MAX) - int(JOYSTICK_AXIS_MIN)
			distance := clampF32((y-virtualAxisStartY)/image.AxisHeight, 0.0, 1.0)
			value := int16(int(JOYSTICK_AXIS_MIN) + int(distance*float32(axisRange)))
			virtualJoystick.SetVirtualAxis(int32(virtualAxisActive), value)
		} else {
			distanceX := clampF32((x-virtualAxisStartX)/image.AxisWidth, -1.0, 1.0)
			distanceY := clampF32((y-virtualAxisStartY)/image.AxisHeight, -1.0, 1.0)

			var valueX, valueY int16
			if distanceX >= 0 {
				valueX = int16(distanceX * float32(JOYSTICK_AXIS_MAX))
			} else {
				valueX = int16(distanceX * float32(-JOYSTICK_AXIS_MIN))
			}
			if distanceY >= 0 {
				valueY = int16(distanceY * float32(JOYSTICK_AXIS_MAX))
			} else {
				valueY = int16(distanceY * float32(-JOYSTICK_AXIS_MIN))
			}
			virtualJoystick.SetVirtualAxis(int32(virtualAxisActive), valueX)
			virtualJoystick.SetVirtualAxis(int32(virtualAxisActive)+1, valueY)
		}
	}

	if virtualTouchpadActive {
		touchpad := image.GetTouchpadArea()
		virtualTouchpadX = (x - touchpad.X) / touchpad.W
		virtualTouchpadY = (y - touchpad.Y) / touchpad.H
		virtualJoystick.SetVirtualTouchpad(0, 0, true, virtualTouchpadX, virtualTouchpadY, 1.0)
	}
}

func VirtualGamepadMouseDown(x, y float32) {
	element := image.GetElementAt(x, y)

	if element == -1 {
		touchpad := image.GetTouchpadArea()
		point := sdl.FPoint{X: x, Y: y}
		if gamepadutils.PointInRectFloat(&point, &touchpad) {
			virtualTouchpadActive = true
			virtualTouchpadX = (x - touchpad.X) / touchpad.W
			virtualTouchpadY = (y - touchpad.Y) / touchpad.H
			virtualJoystick.SetVirtualTouchpad(0, 0, true, virtualTouchpadX, virtualTouchpadY, 1.0)
		}
		return
	}

	if element < int(sdl.GAMEPAD_BUTTON_COUNT) {
		virtualButtonActive = sdl.GamepadButton(element)
		virtualJoystick.SetVirtualButton(int32(virtualButtonActive), true)
	} else {
		switch element {
		case int(sdl.GAMEPAD_BUTTON_COUNT) + 0, // AXIS_LEFTX_NEGATIVE
			int(sdl.GAMEPAD_BUTTON_COUNT) + 1,  // AXIS_LEFTX_POSITIVE
			int(sdl.GAMEPAD_BUTTON_COUNT) + 2,  // AXIS_LEFTY_NEGATIVE
			int(sdl.GAMEPAD_BUTTON_COUNT) + 3:   // AXIS_LEFTY_POSITIVE
			virtualAxisActive = sdl.GAMEPAD_AXIS_LEFTX
		case int(sdl.GAMEPAD_BUTTON_COUNT) + 4, // AXIS_RIGHTX_NEGATIVE
			int(sdl.GAMEPAD_BUTTON_COUNT) + 5,  // AXIS_RIGHTX_POSITIVE
			int(sdl.GAMEPAD_BUTTON_COUNT) + 6,  // AXIS_RIGHTY_NEGATIVE
			int(sdl.GAMEPAD_BUTTON_COUNT) + 7:   // AXIS_RIGHTY_POSITIVE
			virtualAxisActive = sdl.GAMEPAD_AXIS_RIGHTX
		case int(sdl.GAMEPAD_BUTTON_COUNT) + 8: // AXIS_LEFT_TRIGGER
			virtualAxisActive = sdl.GAMEPAD_AXIS_LEFT_TRIGGER
		case int(sdl.GAMEPAD_BUTTON_COUNT) + 9: // AXIS_RIGHT_TRIGGER
			virtualAxisActive = sdl.GAMEPAD_AXIS_RIGHT_TRIGGER
		}
		virtualAxisStartX = x
		virtualAxisStartY = y
	}
}

func VirtualGamepadMouseUp(x, y float32) {
	if virtualButtonActive != sdl.GAMEPAD_BUTTON_INVALID {
		virtualJoystick.SetVirtualButton(int32(virtualButtonActive), false)
		virtualButtonActive = sdl.GAMEPAD_BUTTON_INVALID
	}

	if virtualAxisActive != sdl.GAMEPAD_AXIS_INVALID {
		if virtualAxisActive == sdl.GAMEPAD_AXIS_LEFT_TRIGGER ||
			virtualAxisActive == sdl.GAMEPAD_AXIS_RIGHT_TRIGGER {
			virtualJoystick.SetVirtualAxis(int32(virtualAxisActive), int16(JOYSTICK_AXIS_MIN))
		} else {
			virtualJoystick.SetVirtualAxis(int32(virtualAxisActive), 0)
			virtualJoystick.SetVirtualAxis(int32(virtualAxisActive)+1, 0)
		}
		virtualAxisActive = sdl.GAMEPAD_AXIS_INVALID
	}

	if virtualTouchpadActive {
		virtualJoystick.SetVirtualTouchpad(0, 0, false, virtualTouchpadX, virtualTouchpadY, 0.0)
		virtualTouchpadActive = false
	}
}

func clampF32(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
