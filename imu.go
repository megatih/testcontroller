package main

import (
	"github.com/Zyko0/go-sdl3/sdl"

	"github.com/megatih/testcontroller/gamepadutils"
)

// IMU timing constants.
const (
	NS_PER_SECOND uint64 = 1_000_000_000

	SDL_GAMEPAD_IMU_NOISE_SETTLING_PERIOD_NS            = NS_PER_SECOND / 2
	SDL_GAMEPAD_IMU_NOISE_EVALUATION_PERIOD_NS          = 4 * NS_PER_SECOND
	SDL_GAMEPAD_IMU_NOISE_PROFILING_PHASE_DURATION_NS   = SDL_GAMEPAD_IMU_NOISE_SETTLING_PERIOD_NS + SDL_GAMEPAD_IMU_NOISE_EVALUATION_PERIOD_NS
	SDL_GAMEPAD_IMU_CALIBRATION_PHASE_DURATION_NS       = 5 * NS_PER_SECOND
	SDL_GAMEPAD_IMU_MIN_POLLING_RATE_ESTIMATION_TIME_NS = NS_PER_SECOND * 2
)

// IMUState holds per-controller IMU state.
type IMUState struct {
	GyroPacketNumber         uint64
	AccelerometerPacketNumber uint64
	IMUPacketCounter         uint64

	StartingTimeStampNS    uint64
	IMUEstimatedSensorRate uint16

	LastSensorTimeStampNS uint64

	AccelData     [3]float32
	GyroData      [3]float32
	LastAccelData [3]float32

	AccelerometerLengthSquared    float32
	AccelerometerToleranceSquared float32

	GyroDriftAccumulator [3]float32

	CalibrationPhase              gamepadutils.EGyroCalibrationPhase
	CalibrationPhaseStartTimeNS   uint64

	GyroDriftSampleCount int
	GyroDriftSolution    [3]float32

	IntegratedRotation Quaternion
}

func BeginNoiseCalibrationPhase(imustate *IMUState) {
	imustate.AccelerometerToleranceSquared = gamepadutils.ACCELEROMETER_NOISE_THRESHOLD
	imustate.CalibrationPhase = gamepadutils.GYRO_CALIBRATION_PHASE_NOISE_PROFILING
	imustate.CalibrationPhaseStartTimeNS = sdl.TicksNS()
}

func BeginDriftCalibrationPhase(imustate *IMUState) {
	imustate.CalibrationPhase = gamepadutils.GYRO_CALIBRATION_PHASE_DRIFT_PROFILING
	imustate.CalibrationPhaseStartTimeNS = sdl.TicksNS()
	imustate.GyroDriftSampleCount = 0
	imustate.GyroDriftSolution = [3]float32{}
	imustate.GyroDriftAccumulator = [3]float32{}
}

func ResetIMUState(imustate *IMUState) {
	imustate.GyroPacketNumber = 0
	imustate.AccelerometerPacketNumber = 0
	imustate.StartingTimeStampNS = sdl.TicksNS()
	imustate.IntegratedRotation = QuatIdentity
	imustate.AccelerometerLengthSquared = 0
	imustate.AccelerometerToleranceSquared = gamepadutils.ACCELEROMETER_NOISE_THRESHOLD
	imustate.CalibrationPhase = gamepadutils.GYRO_CALIBRATION_PHASE_OFF
	imustate.CalibrationPhaseStartTimeNS = sdl.TicksNS()
	imustate.LastAccelData = [3]float32{}
	imustate.GyroDriftSolution = [3]float32{}
	imustate.GyroDriftAccumulator = [3]float32{}
}

func ResetGyroOrientation(imustate *IMUState) {
	imustate.IntegratedRotation = QuatIdentity
}

func CalibrationPhaseNoiseProfiling(imustate *IMUState) {
	if imustate.AccelerometerLengthSquared > gamepadutils.ACCELEROMETER_MAX_NOISE_G_SQ {
		BeginNoiseCalibrationPhase(imustate)
		return
	}

	now := sdl.TicksNS()
	deltaNS := now - imustate.CalibrationPhaseStartTimeNS

	if deltaNS > SDL_GAMEPAD_IMU_NOISE_SETTLING_PERIOD_NS {
		if imustate.AccelerometerLengthSquared > imustate.AccelerometerToleranceSquared {
			imustate.AccelerometerToleranceSquared = imustate.AccelerometerLengthSquared
		}
	}

	if deltaNS >= SDL_GAMEPAD_IMU_NOISE_PROFILING_PHASE_DURATION_NS {
		BeginDriftCalibrationPhase(imustate)
	}
}

func FinalizeDriftSolution(imustate *IMUState) {
	if imustate.GyroDriftSampleCount > 0 {
		imustate.GyroDriftSolution[0] = imustate.GyroDriftAccumulator[0] / float32(imustate.GyroDriftSampleCount)
		imustate.GyroDriftSolution[1] = imustate.GyroDriftAccumulator[1] / float32(imustate.GyroDriftSampleCount)
		imustate.GyroDriftSolution[2] = imustate.GyroDriftAccumulator[2] / float32(imustate.GyroDriftSampleCount)
	}

	imustate.CalibrationPhase = gamepadutils.GYRO_CALIBRATION_PHASE_COMPLETE
	ResetGyroOrientation(imustate)
}

func CalibrationPhaseDriftProfiling(imustate *IMUState) {
	if imustate.AccelerometerLengthSquared > imustate.AccelerometerToleranceSquared {
		BeginDriftCalibrationPhase(imustate)
	} else {
		imustate.GyroDriftSampleCount++

		imustate.GyroDriftAccumulator[0] += imustate.GyroData[0]
		imustate.GyroDriftAccumulator[1] += imustate.GyroData[1]
		imustate.GyroDriftAccumulator[2] += imustate.GyroData[2]

		now := sdl.TicksNS()
		deltaNS := now - imustate.CalibrationPhaseStartTimeNS
		if deltaNS >= SDL_GAMEPAD_IMU_CALIBRATION_PHASE_DURATION_NS {
			FinalizeDriftSolution(imustate)
		}
	}
}

func SampleGyroPacketForDrift(imustate *IMUState) {
	diff := [3]float32{
		imustate.AccelData[0] - imustate.LastAccelData[0],
		imustate.AccelData[1] - imustate.LastAccelData[1],
		imustate.AccelData[2] - imustate.LastAccelData[2],
	}
	imustate.LastAccelData = imustate.AccelData
	imustate.AccelerometerLengthSquared = diff[0]*diff[0] + diff[1]*diff[1] + diff[2]*diff[2]

	if imustate.CalibrationPhase == gamepadutils.GYRO_CALIBRATION_PHASE_NOISE_PROFILING {
		CalibrationPhaseNoiseProfiling(imustate)
	}

	if imustate.CalibrationPhase == gamepadutils.GYRO_CALIBRATION_PHASE_DRIFT_PROFILING {
		CalibrationPhaseDriftProfiling(imustate)
	}
}

func ApplyDriftSolution(gyroData *[3]float32, driftSolution [3]float32) {
	gyroData[0] -= driftSolution[0]
	gyroData[1] -= driftSolution[1]
	gyroData[2] -= driftSolution[2]
}

func UpdateGyroRotation(imustate *IMUState, sensorTimeStampDeltaNS uint64) {
	sensorTimeDeltaSeconds := float32(sensorTimeStampDeltaNS) / 1e9
	pitch := imustate.GyroData[0] * sensorTimeDeltaSeconds
	yaw := imustate.GyroData[1] * sensorTimeDeltaSeconds
	roll := imustate.GyroData[2] * sensorTimeDeltaSeconds

	deltaRotation := QuaternionFromEuler(pitch, yaw, roll)
	imustate.IntegratedRotation = MultiplyQuaternion(imustate.IntegratedRotation, deltaRotation)
	NormalizeQuaternion(&imustate.IntegratedRotation)
}

func EstimatePacketRate() {
	if controller == nil || controller.IMUState == nil {
		return
	}

	nowNS := sdl.TicksNS()
	if controller.IMUState.IMUPacketCounter == 0 {
		controller.IMUState.StartingTimeStampNS = nowNS
	}

	deltaTimeNS := nowNS - controller.IMUState.StartingTimeStampNS
	if deltaTimeNS >= SDL_GAMEPAD_IMU_MIN_POLLING_RATE_ESTIMATION_TIME_NS {
		controller.IMUState.IMUEstimatedSensorRate = uint16((controller.IMUState.IMUPacketCounter * NS_PER_SECOND) / deltaTimeNS)
		controller.IMUState.IMUPacketCounter = 0
	} else {
		controller.IMUState.IMUPacketCounter++
	}
}

func UpdateGamepadOrientation(deltaTimeNS uint64) {
	if controller == nil || controller.IMUState == nil {
		return
	}

	SampleGyroPacketForDrift(controller.IMUState)
	ApplyDriftSolution(&controller.IMUState.GyroData, controller.IMUState.GyroDriftSolution)
	UpdateGyroRotation(controller.IMUState, deltaTimeNS)
}
