package main

import (
	"fmt"
	"log"

	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/megatih/testcontroller/gamepadutils"
)

// guidToString converts an SDL GUID to a hex string.
func guidToString(guid sdl.GUID) string {
	if guid == nil {
		return ""
	}
	b := *guid
	return fmt.Sprintf("%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x%02x",
		b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7],
		b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15])
}

// AxisState tracks axis movement for binding detection.
type AxisState struct {
	Moving        bool
	LastValue     int
	StartingValue int
	FarthestValue int
}

// Controller holds state for a connected controller.
type Controller struct {
	ID sdl.JoystickID

	Joystick  *sdl.Joystick
	NumAxes   int
	AxisStates []AxisState
	IMUState  *IMUState

	Gamepad     *sdl.Gamepad
	Mapping     string
	HasBindings bool

	AudioRoute    int
	TriggerEffect int
}

// DS5EffectsState_t represents the PS5 DualSense effects data structure.
// See: https://controllers.fandom.com/wiki/Sony_DualSense#FFB_Trigger_Modes
type DS5EffectsState struct {
	data [47]byte
}

func (e *DS5EffectsState) setEnableBits1(v byte)                   { e.data[0] |= v }
func (e *DS5EffectsState) setEnableBits2(v byte)                   { e.data[1] |= v }
func (e *DS5EffectsState) setRumbleRight(v byte)                   { e.data[2] = v }
func (e *DS5EffectsState) setRumbleLeft(v byte)                    { e.data[3] = v }
func (e *DS5EffectsState) setHeadphoneVolume(v byte)               { e.data[4] = v }
func (e *DS5EffectsState) setSpeakerVolume(v byte)                 { e.data[5] = v }
func (e *DS5EffectsState) setMicrophoneVolume(v byte)              { e.data[6] = v }
func (e *DS5EffectsState) setAudioEnableBits(v byte)               { e.data[7] = v }
func (e *DS5EffectsState) setRightTriggerEffect(data [11]byte)     { copy(e.data[10:21], data[:]) }
func (e *DS5EffectsState) setLeftTriggerEffect(data [11]byte)      { copy(e.data[21:32], data[:]) }

func CyclePS5AudioRoute(device *Controller) {
	device.AudioRoute = (device.AudioRoute + 1) % 4

	var effects DS5EffectsState
	switch device.AudioRoute {
	case 0:
		effects.setEnableBits1(0x80 | 0x20 | 0x10)
		effects.setSpeakerVolume(0)
		effects.setHeadphoneVolume(0)
		effects.setAudioEnableBits(0x00)
	case 1:
		effects.setEnableBits1(0x80 | 0x10)
		effects.setHeadphoneVolume(50)
		effects.setAudioEnableBits(0x00)
	case 2:
		effects.setEnableBits1(0x80 | 0x20)
		effects.setSpeakerVolume(100)
		effects.setAudioEnableBits(0x30)
	case 3:
		effects.setEnableBits1(0x80 | 0x20 | 0x10)
		effects.setSpeakerVolume(100)
		effects.setHeadphoneVolume(50)
		effects.setAudioEnableBits(0x20)
	}
	device.Gamepad.SendEffect(effects.data[:])
}

func CyclePS5TriggerEffect(device *Controller) {
	triggerEffects := [3][11]byte{
		{0x05, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0x01, 0, 110, 0, 0, 0, 0, 0, 0, 0, 0},
		{0x06, 15, 63, 128, 0, 0, 0, 0, 0, 0, 0},
	}

	device.TriggerEffect = (device.TriggerEffect + 1) % len(triggerEffects)

	var effects DS5EffectsState
	effects.setEnableBits1(0x04 | 0x08)
	effects.setRightTriggerEffect(triggerEffects[device.TriggerEffect])
	effects.setLeftTriggerEffect(triggerEffects[device.TriggerEffect])
	device.Gamepad.SendEffect(effects.data[:])
}

func FindController(id sdl.JoystickID) int {
	for i := 0; i < numControllers; i++ {
		if id == controllers[i].ID {
			return i
		}
	}
	return -1
}

func SetController(id sdl.JoystickID) {
	i := FindController(id)
	if i < 0 && numControllers > 0 {
		i = 0
	}
	if i >= 0 {
		controller = &controllers[i]
	} else {
		controller = nil
	}
	RefreshControllerName()
}

func AddController(id sdl.JoystickID, verbose bool) {
	if FindController(id) >= 0 {
		return
	}

	controllers = append(controllers, Controller{})
	controller = nil
	numControllers = len(controllers)
	newController := &controllers[numControllers-1]
	newController.ID = id

	joystick, err := id.OpenJoystick()
	if err != nil {
		log.Printf("Couldn't open joystick: %v", err)
	}
	newController.Joystick = joystick

	if joystick != nil {
		numAxes, _ := joystick.NumAxes()
		newController.NumAxes = int(numAxes)
		newController.AxisStates = make([]AxisState, newController.NumAxes)
		newController.IMUState = &IMUState{}
		ResetIMUState(newController.IMUState)

		if verbose && !id.IsGamepad() {
			name, _ := joystick.Name()
			path, _ := joystick.Path()
			log.Printf("Opened joystick %s%s%s", name,
				func() string {
					if path != "" {
						return ", "
					}
					return ""
				}(), path)
			guid := joystick.GUID()
			log.Printf("No gamepad mapping for %s", guidToString(guid))
		}
	}

	if mappingController != 0 {
		SetController(mappingController)
	} else {
		SetController(id)
	}
}

func DelController(id sdl.JoystickID) {
	i := FindController(id)
	if i < 0 {
		return
	}

	if displayMode == gamepadutils.CONTROLLER_MODE_BINDING && controller != nil && id == controller.ID {
		SetDisplayMode(gamepadutils.CONTROLLER_MODE_TESTING)
	}

	if controllers[i].TriggerEffect != 0 {
		controllers[i].TriggerEffect = -1
		CyclePS5TriggerEffect(&controllers[i])
	}

	if controllers[i].Joystick != nil {
		controllers[i].Joystick.Close()
	}

	controllers = append(controllers[:i], controllers[i+1:]...)
	numControllers = len(controllers)

	if mappingController != 0 {
		SetController(mappingController)
	} else {
		SetController(id)
	}
}

func HandleGamepadRemapped(id sdl.JoystickID) {
	i := FindController(id)
	if i < 0 {
		return
	}

	if controllers[i].Gamepad == nil {
		return
	}

	mapping, err := controllers[i].Gamepad.Mapping()
	if err != nil {
		mapping = ""
	}

	if mapping != "" && !gamepadutils.MappingHasName(mapping) {
		name, _ := controllers[i].Joystick.Name()
		mapping = gamepadutils.SetMappingName(mapping, name)
	}

	controllers[i].Mapping = mapping
	controllers[i].HasBindings = gamepadutils.MappingHasBindings(mapping)
}

func HandleGamepadAdded(id sdl.JoystickID, verbose bool) {
	sensors := []sdl.SensorType{
		sdl.SENSOR_ACCEL,
		sdl.SENSOR_GYRO,
		sdl.SENSOR_ACCEL_L,
		sdl.SENSOR_GYRO_L,
		sdl.SENSOR_ACCEL_R,
		sdl.SENSOR_GYRO_R,
	}

	i := FindController(id)
	if i < 0 {
		return
	}
	log.Printf("Gamepad %d added", id)

	gamepad, err := id.OpenGamepad()
	if err != nil {
		log.Printf("Couldn't open gamepad: %v", err)
		HandleGamepadRemapped(id)
		SetController(id)
		return
	}
	controllers[i].Gamepad = gamepad

	if verbose {
		name := gamepad.Name()
		path := gamepad.Path()
		guid := guidToString(id.JoystickGUIDForID())
		pathStr := ""
		if path != "" {
			pathStr = ", " + path
		}
		log.Printf("Opened gamepad %s, guid %s%s", name, guid, pathStr)

		fw := gamepad.FirmwareVersion()
		if fw != 0 {
			log.Printf("Firmware version: 0x%x (%d)", fw, fw)
		}

		if gamepad.PlayerIndex() >= 0 {
			log.Printf("Player index: %d", gamepad.PlayerIndex())
		}
	}

	for _, sensor := range sensors {
		if gamepad.HasSensor(sensor) {
			if verbose {
				log.Printf("Enabling %s at %.2f Hz", getSensorName(sensor), gamepad.SensorDataRate(sensor))
			}
			gamepad.SetSensorEnabled(sensor, true)
		}
	}

	if verbose {
		mapping, err := gamepad.Mapping()
		if err == nil && mapping != "" {
			log.Printf("Mapping: %s", mapping)
		}
	}

	HandleGamepadRemapped(id)
	SetController(id)
}

func HandleGamepadRemoved(id sdl.JoystickID) {
	i := FindController(id)
	if i < 0 {
		return
	}
	log.Printf("Gamepad %d removed", id)

	controllers[i].Mapping = ""
	if controllers[i].Gamepad != nil {
		controllers[i].Gamepad.Close()
		controllers[i].Gamepad = nil
	}
}

func HandleGamepadSensorEvent(event *sdl.Event) {
	if controller == nil {
		return
	}

	evt := event.GamepadSensorEvent()
	if controller.ID != evt.Which {
		return
	}

	sensor := sdl.SensorType(evt.Sensor)
	if sensor == sdl.SENSOR_GYRO {
		controller.IMUState.GyroPacketNumber++
		controller.IMUState.GyroData = evt.Data
	} else if sensor == sdl.SENSOR_ACCEL {
		controller.IMUState.AccelerometerPacketNumber++
		controller.IMUState.AccelData = evt.Data
	}

	if controller.IMUState.AccelerometerPacketNumber == controller.IMUState.GyroPacketNumber {
		EstimatePacketRate()
		sensorTimeStampDeltaNS := evt.SensorTimestamp - controller.IMUState.LastSensorTimeStampNS
		UpdateGamepadOrientation(sensorTimeStampDeltaNS)

		pitch, yaw, roll := QuaternionToYXZ(controller.IMUState.IntegratedRotation)
		displayEulerAngles := [3]float32{pitch, yaw, roll}

		now := sdl.TicksNS()
		var duration uint64
		if controller.IMUState.CalibrationPhase == gamepadutils.GYRO_CALIBRATION_PHASE_NOISE_PROFILING {
			duration = SDL_GAMEPAD_IMU_NOISE_PROFILING_PHASE_DURATION_NS
		} else if controller.IMUState.CalibrationPhase == gamepadutils.GYRO_CALIBRATION_PHASE_DRIFT_PROFILING {
			duration = SDL_GAMEPAD_IMU_CALIBRATION_PHASE_DURATION_NS
		}

		deltaNS := now - controller.IMUState.CalibrationPhaseStartTimeNS
		var driftCalibrationProgressFrac float32
		if duration > 0 {
			driftCalibrationProgressFrac = float32(deltaNS) / float32(duration)
		}

		var reportedPollingRateHz int
		if sensorTimeStampDeltaNS > 0 {
			reportedPollingRateHz = int(NS_PER_SECOND / sensorTimeStampDeltaNS)
		}

		gyroElements.SetIMUValues(
			controller.IMUState.GyroDriftSolution,
			displayEulerAngles,
			(*gamepadutils.Quaternion)(&controller.IMUState.IntegratedRotation),
			reportedPollingRateHz,
			int(controller.IMUState.IMUEstimatedSensorRate),
			controller.IMUState.CalibrationPhase,
			driftCalibrationProgressFrac,
			controller.IMUState.AccelerometerLengthSquared,
			controller.IMUState.AccelerometerToleranceSquared,
		)

		gamepadElements.SetGyroDriftCorrection(controller.IMUState.GyroDriftSolution)
		controller.IMUState.LastSensorTimeStampNS = evt.SensorTimestamp
	}
}

func RefreshControllerName() {
	controllerName = ""
	if controller != nil {
		if controller.Gamepad != nil {
			controllerName = controller.Gamepad.Name()
		} else if controller.Joystick != nil {
			name, _ := controller.Joystick.Name()
			controllerName = name
		}
	}
}
