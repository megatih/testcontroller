package gamepadutils

import (
	"fmt"
	"math"

	"github.com/Zyko0/go-sdl3/sdl"
)

// GyroDisplay shows the gyro/IMU visualization with calibration controls.
type GyroDisplay struct {
	renderer *sdl.Renderer

	Area sdl.FRect

	GyroDriftSolution      [3]float32
	ReportedSensorRateHz   int
	nextReportedSensorTime uint64

	EstimatedSensorRateHz            int
	EulerDisplacementAngles          [3]float32
	GyroQuaternion                   Quaternion
	CurrentCalibrationPhase          EGyroCalibrationPhase
	CalibrationPhaseProgressFraction float32
	AccelerometerNoiseSq             float32
	AccelerometerNoiseToleranceSq    float32

	ResetGyroButton     *GamepadButton
	CalibrateGyroButton *GamepadButton
}

// CreateGyroDisplay creates a new gyro display widget.
func CreateGyroDisplay(renderer *sdl.Renderer) *GyroDisplay {
	ctx := &GyroDisplay{
		renderer:                      renderer,
		GyroQuaternion:                Quaternion{X: 0, Y: 0, Z: 0, W: 1},
		CurrentCalibrationPhase:       GYRO_CALIBRATION_PHASE_OFF,
		AccelerometerNoiseToleranceSq: ACCELEROMETER_NOISE_THRESHOLD,
		ResetGyroButton:               CreateGamepadButton(renderer, "Reset View"),
		CalibrateGyroButton:           CreateGamepadButton(renderer, "Recalibrate Drift"),
	}
	return ctx
}

// SetArea sets the gyro display area and positions the reset button.
func (ctx *GyroDisplay) SetArea(area *sdl.FRect) {
	if ctx == nil {
		return
	}
	ctx.Area = *area

	resetArea := sdl.FRect{
		W: max32(MINIMUM_BUTTON_WIDTH, ctx.ResetGyroButton.LabelWidth()+2*BUTTON_PADDING),
		H: ctx.ResetGyroButton.LabelHeight() + BUTTON_PADDING,
	}
	resetArea.X = area.X + area.W - resetArea.W - BUTTON_PADDING
	resetArea.Y = area.Y + area.H - resetArea.H - BUTTON_PADDING
	ctx.ResetGyroButton.SetArea(&resetArea)
}

// SetIMUValues updates all the IMU display values.
func (ctx *GyroDisplay) SetIMUValues(gyroDriftSolution [3]float32, eulerAngles [3]float32, gyroQuat *Quaternion, reportedRateHz, estimatedRateHz int, calibrationPhase EGyroCalibrationPhase, driftProgress float32, accelNoiseSq, accelNoiseTolSq float32) {
	if ctx == nil {
		return
	}

	const sensorUpdateIntervalMS = 100
	now := sdl.Ticks()
	if now > ctx.nextReportedSensorTime {
		ctx.EstimatedSensorRateHz = estimatedRateHz
		if reportedRateHz != 0 {
			ctx.ReportedSensorRateHz = reportedRateHz
		}
		ctx.nextReportedSensorTime = now + sensorUpdateIntervalMS
	}

	ctx.GyroDriftSolution = gyroDriftSolution
	ctx.EulerDisplacementAngles = eulerAngles
	ctx.GyroQuaternion = *gyroQuat
	ctx.CurrentCalibrationPhase = calibrationPhase
	ctx.CalibrationPhaseProgressFraction = driftProgress
	ctx.AccelerometerNoiseSq = accelNoiseSq
	ctx.AccelerometerNoiseToleranceSq = accelNoiseTolSq
}

// renderSensorTimingInfo draws the sensor timing section.
func (ctx *GyroDisplay) renderSensorTimingInfo(gamepadDisplay *GamepadDisplay) {
	newLineHeight := gamepadDisplay.ButtonHeight + 2.0
	textOffsetX := ctx.Area.X + ctx.Area.W/4.0 + 35.0
	textYPos := ctx.Area.Y + ctx.Area.H - newLineHeight*2

	text := "HID Sensor Time:"
	ctx.renderer.DebugText(textOffsetX-float32(len(text))*FONT_CHARACTER_SIZE, textYPos, text)
	if ctx.ReportedSensorRateHz > 0 {
		deltaTimeUs := int(1e6) / ctx.ReportedSensorRateHz
		text = fmt.Sprintf("%d%ss %dhz", deltaTimeUs, MICRO_UTF8, ctx.ReportedSensorRateHz)
	} else {
		text = fmt.Sprintf("????%ss ???hz", MICRO_UTF8)
	}
	ctx.renderer.DebugText(textOffsetX+2.0, textYPos, text)

	textYPos += newLineHeight
	text = "Est.Sensor Time:"
	ctx.renderer.DebugText(textOffsetX-float32(len(text))*FONT_CHARACTER_SIZE, textYPos, text)
	if ctx.EstimatedSensorRateHz > 0 {
		deltaTimeUs := int(1e6) / ctx.EstimatedSensorRateHz
		text = fmt.Sprintf("%d%ss %dhz", deltaTimeUs, MICRO_UTF8, ctx.EstimatedSensorRateHz)
	} else {
		text = fmt.Sprintf("????%ss ???hz", MICRO_UTF8)
	}
	ctx.renderer.DebugText(textOffsetX+2.0, textYPos, text)
}

// renderDriftCalibrationButton draws the calibration button and progress bars.
func (ctx *GyroDisplay) renderDriftCalibrationButton(gamepadDisplay *GamepadDisplay) {
	newLineHeight := gamepadDisplay.ButtonHeight + 2.0
	logY := ctx.Area.Y + BUTTON_PADDING

	// Position the calibrate button
	recalibrateButtonWidth := ctx.CalibrateGyroButton.LabelWidth() + 2*BUTTON_PADDING
	recalibArea := sdl.FRect{
		X: ctx.Area.X + ctx.Area.W - recalibrateButtonWidth - BUTTON_PADDING,
		Y: logY + FONT_CHARACTER_SIZE*0.5 - gamepadDisplay.ButtonHeight*0.5,
		W: ctx.CalibrateGyroButton.LabelWidth() + 2.0*BUTTON_PADDING,
		H: gamepadDisplay.ButtonHeight + BUTTON_PADDING*2.0,
	}

	// Label above button
	ctx.renderer.DebugText(recalibArea.X, recalibArea.Y-newLineHeight, "Gyro Orientation:")

	// Update button label based on calibration phase
	var labelText string
	switch ctx.CurrentCalibrationPhase {
	case GYRO_CALIBRATION_PHASE_OFF:
		labelText = "Start Gyro Calibration"
	case GYRO_CALIBRATION_PHASE_NOISE_PROFILING:
		labelText = fmt.Sprintf("Noise Progress: %3.0f%% ", ctx.CalibrationPhaseProgressFraction*100.0)
	case GYRO_CALIBRATION_PHASE_DRIFT_PROFILING:
		labelText = fmt.Sprintf("Drift Progress: %3.0f%% ", ctx.CalibrationPhaseProgressFraction*100.0)
	case GYRO_CALIBRATION_PHASE_COMPLETE:
		labelText = "Recalibrate Gyro"
	}

	ctx.CalibrateGyroButton.SetLabel(labelText)
	ctx.CalibrateGyroButton.SetArea(&recalibArea)
	ctx.CalibrateGyroButton.Render()

	extremeNoise := ctx.AccelerometerNoiseSq > ACCELEROMETER_MAX_NOISE_G_SQ

	if ctx.CurrentCalibrationPhase == GYRO_CALIBRATION_PHASE_OFF {
		if extremeNoise {
			ctx.renderer.DebugText(recalibArea.X, recalibArea.Y+recalibArea.H+newLineHeight, "GamePad Must Be Still")
			ctx.renderer.DebugText(recalibArea.X, recalibArea.Y+recalibArea.H+newLineHeight*2, "Place GamePad On Table")
		}
	}

	if ctx.CurrentCalibrationPhase == GYRO_CALIBRATION_PHASE_NOISE_PROFILING ||
		ctx.CurrentCalibrationPhase == GYRO_CALIBRATION_PHASE_DRIFT_PROFILING {

		absNoiseFrac := clamp32(ctx.AccelerometerNoiseSq/ACCELEROMETER_MAX_NOISE_G_SQ, 0, 1)
		absToleranceFrac := clamp32(ctx.AccelerometerNoiseToleranceSq/ACCELEROMETER_MAX_NOISE_G_SQ, 0, 1)

		maxNoiseForPhase := float32(ACCELEROMETER_MAX_NOISE_G_SQ)
		if ctx.CurrentCalibrationPhase != GYRO_CALIBRATION_PHASE_NOISE_PROFILING {
			maxNoiseForPhase = ctx.AccelerometerNoiseToleranceSq
		}
		relNoiseFrac := clamp32(ctx.AccelerometerNoiseSq/maxNoiseForPhase, 0, 1)

		noiseBarHeight := gamepadDisplay.ButtonHeight
		noiseBarRect := sdl.FRect{
			X: recalibArea.X,
			Y: recalibArea.Y + recalibArea.H + BUTTON_PADDING,
			W: recalibArea.W,
			H: noiseBarHeight,
		}

		ctx.renderer.DebugText(recalibArea.X, recalibArea.Y+recalibArea.H+newLineHeight*2,
			fmt.Sprintf("Accelerometer Noise Tolerance: %3.3fG ", float32(math.Sqrt(float64(ctx.AccelerometerNoiseToleranceSq)))))

		// Noise bar fill
		noiseBarFillWidth := absNoiseFrac * noiseBarRect.W
		noiseBarFillRect := sdl.FRect{
			X: noiseBarRect.X + (noiseBarRect.W-noiseBarFillWidth)*0.5,
			Y: noiseBarRect.Y,
			W: noiseBarFillWidth,
			H: noiseBarHeight,
		}

		red := uint8(relNoiseFrac * 255.0)
		green := uint8((1.0 - relNoiseFrac) * 255.0)
		ctx.renderer.SetDrawColor(red, green, 0, 255)
		ctx.renderer.RenderFillRect(&noiseBarFillRect)

		// Tolerance bar outline
		toleranceBarFillWidth := absToleranceFrac * noiseBarRect.W
		toleranceBarRect := sdl.FRect{
			X: noiseBarRect.X + (noiseBarRect.W-toleranceBarFillWidth)*0.5,
			Y: noiseBarRect.Y,
			W: toleranceBarFillWidth,
			H: noiseBarHeight,
		}
		ctx.renderer.SetDrawColor(128, 128, 0, 255)
		ctx.renderer.RenderRect(&toleranceBarRect)

		ctx.renderer.SetDrawColor(100, 100, 100, 255)
		ctx.renderer.RenderRect(&noiseBarRect)

		tooMuchNoise := absNoiseFrac >= 1.0
		if tooMuchNoise {
			ctx.renderer.DebugText(recalibArea.X, noiseBarRect.Y+noiseBarRect.H+newLineHeight, "Place GamePad Down!")
		}

		// Progress bar
		progressBarRect := sdl.FRect{
			X: recalibArea.X + BUTTON_PADDING,
			Y: recalibArea.Y + recalibArea.H*0.5 + BUTTON_PADDING*0.5,
			W: recalibArea.W - BUTTON_PADDING*2.0,
			H: BUTTON_PADDING * 0.5,
		}

		driftBarFillWidth := ctx.CalibrationPhaseProgressFraction * progressBarRect.W
		if tooMuchNoise {
			driftBarFillWidth = 1.0
		}
		progressBarFill := sdl.FRect{
			X: progressBarRect.X,
			Y: progressBarRect.Y,
			W: driftBarFillWidth,
			H: progressBarRect.H,
		}

		ctx.renderer.SetDrawColor(GYRO_COLOR_GREEN[0], GYRO_COLOR_GREEN[1], GYRO_COLOR_GREEN[2], GYRO_COLOR_GREEN[3])
		ctx.renderer.RenderFillRect(&progressBarFill)
		ctx.renderer.SetDrawColor(100, 100, 100, 255)
		ctx.renderer.RenderRect(&progressBarRect)

		if tooMuchNoise {
			ctx.renderer.SetDrawColor(GYRO_COLOR_RED[0], GYRO_COLOR_RED[1], GYRO_COLOR_RED[2], GYRO_COLOR_RED[3])
			ctx.renderer.RenderFillRect(&progressBarFill)
		}
	}
}

// renderEulerReadout draws the pitch/yaw/roll euler angle readout.
func (ctx *GyroDisplay) renderEulerReadout(gamepadDisplay *GamepadDisplay) float32 {
	btnArea := ctx.CalibrateGyroButton.GetArea()

	newLineHeight := gamepadDisplay.ButtonHeight + 2.0
	logY := btnArea.Y + btnArea.H + BUTTON_PADDING
	logX := btnArea.X

	savedColor := saveColor(ctx.renderer)

	// Pitch
	ctx.renderer.SetDrawColor(GYRO_COLOR_RED[0], GYRO_COLOR_RED[1], GYRO_COLOR_RED[2], GYRO_COLOR_RED[3])
	ctx.renderer.DebugText(logX+2.0, logY, fmt.Sprintf("Pitch: %6.2f%s", ctx.EulerDisplacementAngles[0], DEGREE_UTF8))

	// Yaw
	ctx.renderer.SetDrawColor(GYRO_COLOR_GREEN[0], GYRO_COLOR_GREEN[1], GYRO_COLOR_GREEN[2], GYRO_COLOR_GREEN[3])
	logY += newLineHeight
	ctx.renderer.DebugText(logX+2.0, logY, fmt.Sprintf("  Yaw: %6.2f%s", ctx.EulerDisplacementAngles[1], DEGREE_UTF8))

	// Roll
	ctx.renderer.SetDrawColor(GYRO_COLOR_BLUE[0], GYRO_COLOR_BLUE[1], GYRO_COLOR_BLUE[2], GYRO_COLOR_BLUE[3])
	logY += newLineHeight
	ctx.renderer.DebugText(logX+2.0, logY, fmt.Sprintf(" Roll: %6.2f%s", ctx.EulerDisplacementAngles[2], DEGREE_UTF8))

	restoreColor(ctx.renderer, savedColor)
	return logY + newLineHeight
}

// renderGizmo draws the 3D cube, circles and accel arrow.
func (ctx *GyroDisplay) renderGizmo(gamepad *sdl.Gamepad, top float32) {
	btnArea := ctx.CalibrateGyroButton.GetArea()

	gizmoSize := btnArea.W
	gizmoRect := sdl.FRect{
		X: btnArea.X + (btnArea.W-gizmoSize)*0.5,
		Y: top,
		W: gizmoSize,
		H: gizmoSize,
	}

	DrawGyroDebugCube(ctx.renderer, &ctx.GyroQuaternion, &gizmoRect)
	DrawGyroDebugAxes(ctx.renderer, &ctx.GyroQuaternion, &gizmoRect)
	DrawGyroDebugCircle(ctx.renderer, &ctx.GyroQuaternion, &gizmoRect)

	if gamepad.HasSensor(sdl.SENSOR_ACCEL) {
		var accelData [3]float32
		gamepad.SensorData(sdl.SENSOR_ACCEL, &accelData[0], 3)
		DrawAccelerometerDebugArrow(ctx.renderer, &ctx.GyroQuaternion, accelData[:], &gizmoRect)
	}

	// Position reset button below gizmo
	if ctx.ResetGyroButton != nil {
		resetArea := ctx.ResetGyroButton.GetArea()
		resetArea.X = btnArea.X
		resetArea.Y = gizmoRect.Y + gizmoRect.H + BUTTON_PADDING*0.5
		resetArea.W = btnArea.W
		resetArea.H = btnArea.H
		ctx.ResetGyroButton.SetArea(&resetArea)
		ctx.ResetGyroButton.Render()
	}
}

// Render draws the full gyro display.
func (ctx *GyroDisplay) Render(gamepadDisplay *GamepadDisplay, gamepad *sdl.Gamepad) {
	if ctx == nil {
		return
	}

	hasAccel := gamepad.HasSensor(sdl.SENSOR_ACCEL)
	hasGyro := gamepad.HasSensor(sdl.SENSOR_GYRO)
	if !hasAccel && !hasGyro {
		return
	}

	savedColor := saveColor(ctx.renderer)

	ctx.renderSensorTimingInfo(gamepadDisplay)
	ctx.renderDriftCalibrationButton(gamepadDisplay)

	if ctx.CurrentCalibrationPhase == GYRO_CALIBRATION_PHASE_COMPLETE {
		bottom := ctx.renderEulerReadout(gamepadDisplay)
		ctx.renderGizmo(gamepad, bottom)
	}

	restoreColor(ctx.renderer, savedColor)
}

// Destroy releases the gyro display resources.
func (ctx *GyroDisplay) Destroy() {
	if ctx == nil {
		return
	}
	if ctx.ResetGyroButton != nil {
		ctx.ResetGyroButton.Destroy()
	}
	if ctx.CalibrateGyroButton != nil {
		ctx.CalibrateGyroButton.Destroy()
	}
}

// Helper functions

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func clamp32(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
