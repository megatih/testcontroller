package gamepadutils

import (
	"math"

	"github.com/Zyko0/go-sdl3/sdl"
)

// RAD_TO_DEG converts radians to degrees.
const RAD_TO_DEG_F = float32(180.0 / math.Pi)

// Vector3 represents a 3D point/vector.
type Vector3 struct {
	X, Y, Z float32
}

var debugCubeVertices = [8]Vector3{
	{-1, -1, -1},
	{1, -1, -1},
	{1, 1, -1},
	{-1, 1, -1},
	{-1, -1, 1},
	{1, -1, 1},
	{1, 1, 1},
	{-1, 1, 1},
}

var debugCubeEdges = [12][2]int{
	{0, 1}, {1, 2}, {2, 3}, {3, 0}, // bottom square
	{4, 5}, {5, 6}, {6, 7}, {7, 4}, // top square
	{0, 4}, {1, 5}, {2, 6}, {3, 7}, // verticals
}

// RotateVectorByQuaternion rotates a vector by a quaternion (v' = q * v * q^-1).
func RotateVectorByQuaternion(v *Vector3, q *Quaternion) Vector3 {
	x, y, z := v.X, v.Y, v.Z
	qx, qy, qz, qw := q.X, q.Y, q.Z, q.W

	// Calculate quaternion * vector
	ix := qw*x + qy*z - qz*y
	iy := qw*y + qz*x - qx*z
	iz := qw*z + qx*y - qy*x
	iw := -qx*x - qy*y - qz*z

	// Result = result * conjugate(q)
	return Vector3{
		X: ix*qw + iw*(-qx) + iy*(-qz) - iz*(-qy),
		Y: iy*qw + iw*(-qy) + iz*(-qx) - ix*(-qz),
		Z: iz*qw + iw*(-qz) + ix*(-qy) - iy*(-qx),
	}
}

// ProjectVec3ToRect projects a 3D point onto a 2D rectangle using perspective projection.
func ProjectVec3ToRect(v *Vector3, rect *sdl.FRect) sdl.FPoint {
	const verticalFOVDeg = 40.0
	const cameraZ = 4.0
	aspect := rect.W / rect.H

	fovScaleY := float32(math.Tan(float64((verticalFOVDeg/180.0)*math.Pi) * 0.5))
	fovScaleX := fovScaleY * aspect

	relZ := cameraZ - v.Z
	if relZ < 0.01 {
		relZ = 0.01
	}

	ndcX := (v.X / relZ) / fovScaleX
	ndcY := (v.Y / relZ) / fovScaleY

	return sdl.FPoint{
		X: rect.X + (rect.W/2.0) + (ndcX * rect.W / 2.0),
		Y: rect.Y + (rect.H/2.0) - (ndcY * rect.H / 2.0),
	}
}

const circleSegments = 64

var (
	circlePointsXY [circleSegments]Vector3
	circlePointsXZ [circleSegments]Vector3
	circlePointsYZ [circleSegments]Vector3
)

// InitCirclePoints3D precomputes circle points for gyro visualization.
func InitCirclePoints3D() {
	for i := 0; i < circleSegments; i++ {
		theta := (float64(i) / circleSegments) * math.Pi * 2.0
		cosT := float32(math.Cos(theta))
		sinT := float32(math.Sin(theta))

		circlePointsXY[i] = Vector3{X: cosT, Y: sinT, Z: 0}
		circlePointsXZ[i] = Vector3{X: cosT, Y: 0, Z: sinT}
		circlePointsYZ[i] = Vector3{X: 0, Y: cosT, Z: sinT}
	}
}

// DrawGyroDebugCube draws a wireframe cube rotated by the given quaternion.
func DrawGyroDebugCube(renderer *sdl.Renderer, orientation *Quaternion, rect *sdl.FRect) {
	var projected [8]sdl.FPoint
	for i := 0; i < 8; i++ {
		rotated := RotateVectorByQuaternion(&debugCubeVertices[i], orientation)
		projected[i] = ProjectVec3ToRect(&rotated, rect)
	}

	for i := 0; i < 12; i++ {
		p0 := projected[debugCubeEdges[i][0]]
		p1 := projected[debugCubeEdges[i][1]]
		renderer.RenderLine(p0.X, p0.Y, p1.X, p1.Y)
	}
}

// DrawGyroCircle draws a circle in 3D space rotated by the given quaternion.
func DrawGyroCircle(renderer *sdl.Renderer, circlePoints []Vector3, numSegments int, orientation *Quaternion, bounds *sdl.FRect, r, g, b, a uint8) {
	renderer.SetDrawColor(r, g, b, a)

	var lastPt sdl.FPoint
	hasLast := false
	for i := 0; i <= numSegments; i++ {
		index := i % numSegments
		rotated := RotateVectorByQuaternion(&circlePoints[index], orientation)
		pt := ProjectVec3ToRect(&rotated, bounds)

		if hasLast {
			renderer.RenderLine(lastPt.X, lastPt.Y, pt.X, pt.Y)
		}
		lastPt = pt
		hasLast = true
	}
}

// DrawGyroDebugCircle draws the XYZ circles for gyro visualization.
func DrawGyroDebugCircle(renderer *sdl.Renderer, orientation *Quaternion, bounds *sdl.FRect) {
	savedColor := saveColor(renderer)

	DrawGyroCircle(renderer, circlePointsYZ[:], circleSegments, orientation, bounds,
		GYRO_COLOR_RED[0], GYRO_COLOR_RED[1], GYRO_COLOR_RED[2], GYRO_COLOR_RED[3]) // X axis - pitch
	DrawGyroCircle(renderer, circlePointsXZ[:], circleSegments, orientation, bounds,
		GYRO_COLOR_GREEN[0], GYRO_COLOR_GREEN[1], GYRO_COLOR_GREEN[2], GYRO_COLOR_GREEN[3]) // Y axis - yaw
	DrawGyroCircle(renderer, circlePointsXY[:], circleSegments, orientation, bounds,
		GYRO_COLOR_BLUE[0], GYRO_COLOR_BLUE[1], GYRO_COLOR_BLUE[2], GYRO_COLOR_BLUE[3]) // Z axis - roll

	restoreColor(renderer, savedColor)
}

// DrawGyroDebugAxes draws the positive XYZ axes as lines.
func DrawGyroDebugAxes(renderer *sdl.Renderer, orientation *Quaternion, bounds *sdl.FRect) {
	savedColor := saveColor(renderer)

	origin := Vector3{0, 0, 0}
	right := Vector3{1, 0, 0}
	up := Vector3{0, 1, 0}
	back := Vector3{0, 0, 1}

	worldRight := RotateVectorByQuaternion(&right, orientation)
	worldUp := RotateVectorByQuaternion(&up, orientation)
	worldBack := RotateVectorByQuaternion(&back, orientation)

	originScreen := ProjectVec3ToRect(&origin, bounds)
	rightScreen := ProjectVec3ToRect(&worldRight, bounds)
	upScreen := ProjectVec3ToRect(&worldUp, bounds)
	backScreen := ProjectVec3ToRect(&worldBack, bounds)

	renderer.SetDrawColor(GYRO_COLOR_RED[0], GYRO_COLOR_RED[1], GYRO_COLOR_RED[2], GYRO_COLOR_RED[3])
	renderer.RenderLine(originScreen.X, originScreen.Y, rightScreen.X, rightScreen.Y)
	renderer.SetDrawColor(GYRO_COLOR_GREEN[0], GYRO_COLOR_GREEN[1], GYRO_COLOR_GREEN[2], GYRO_COLOR_GREEN[3])
	renderer.RenderLine(originScreen.X, originScreen.Y, upScreen.X, upScreen.Y)
	renderer.SetDrawColor(GYRO_COLOR_BLUE[0], GYRO_COLOR_BLUE[1], GYRO_COLOR_BLUE[2], GYRO_COLOR_BLUE[3])
	renderer.RenderLine(originScreen.X, originScreen.Y, backScreen.X, backScreen.Y)

	restoreColor(renderer, savedColor)
}

// DrawAccelerometerDebugArrow draws the accelerometer vector as an arrow.
func DrawAccelerometerDebugArrow(renderer *sdl.Renderer, gyroQuaternion *Quaternion, accelData []float32, bounds *sdl.FRect) {
	savedColor := saveColor(renderer)

	const gravity = 9.81
	vAccel := Vector3{
		X: accelData[0] / gravity,
		Y: accelData[1] / gravity,
		Z: accelData[2] / gravity,
	}

	origin := Vector3{0, 0, 0}
	rotatedAccel := RotateVectorByQuaternion(&vAccel, gyroQuaternion)

	originScreen := ProjectVec3ToRect(&origin, bounds)
	accelScreen := ProjectVec3ToRect(&rotatedAccel, bounds)

	renderer.SetDrawColor(GYRO_COLOR_ORANGE[0], GYRO_COLOR_ORANGE[1], GYRO_COLOR_ORANGE[2], GYRO_COLOR_ORANGE[3])
	renderer.RenderLine(originScreen.X, originScreen.Y, accelScreen.X, accelScreen.Y)

	const headWidth = 4.0
	arrowHead := sdl.FRect{
		X: accelScreen.X - headWidth*0.5,
		Y: accelScreen.Y - headWidth*0.5,
		W: headWidth,
		H: headWidth,
	}
	renderer.RenderRect(&arrowHead)

	restoreColor(renderer, savedColor)
}
