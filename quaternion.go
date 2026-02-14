package main

import (
	"math"

	"github.com/megatih/testcontroller/gamepadutils"
)

type Quaternion = gamepadutils.Quaternion

var QuatIdentity = Quaternion{X: 0, Y: 0, Z: 0, W: 1}

func QuaternionFromEuler(pitch, yaw, roll float32) Quaternion {
	cx := float32(math.Cos(float64(pitch * 0.5)))
	sx := float32(math.Sin(float64(pitch * 0.5)))
	cy := float32(math.Cos(float64(yaw * 0.5)))
	sy := float32(math.Sin(float64(yaw * 0.5)))
	cz := float32(math.Cos(float64(roll * 0.5)))
	sz := float32(math.Sin(float64(roll * 0.5)))

	return Quaternion{
		W: cx*cy*cz + sx*sy*sz,
		X: sx*cy*cz - cx*sy*sz,
		Y: cx*sy*cz + sx*cy*sz,
		Z: cx*cy*sz - sx*sy*cz,
	}
}

const RAD_TO_DEG = float32(180.0 / math.Pi)

// QuaternionToYXZ decomposes quaternion into Yaw (Y), Pitch (X), Roll (Z)
// using Y-X-Z order in a left-handed system.
func QuaternionToYXZ(q Quaternion) (pitch, yaw, roll float32) {
	qxx := q.X * q.X
	qyy := q.Y * q.Y
	qzz := q.Z * q.Z

	qxy := q.X * q.Y
	qxz := q.X * q.Z
	_ = qxy // used below
	_ = qxz
	qyz := q.Y * q.Z
	qwx := q.W * q.X
	qwy := q.W * q.Y
	qwz := q.W * q.Z

	// Yaw (around Y)
	yaw = float32(math.Atan2(float64(2.0*(qwy+qxz)), float64(1.0-2.0*(qyy+qzz)))) * RAD_TO_DEG

	// Pitch (around X)
	sinp := 2.0 * (qwx - qyz)
	if float32(math.Abs(float64(sinp))) >= 1.0 {
		pitch = float32(math.Copysign(90.0, float64(sinp)))
	} else {
		pitch = float32(math.Asin(float64(sinp))) * RAD_TO_DEG
	}

	// Roll (around Z)
	roll = float32(math.Atan2(float64(2.0*(qwz+qxy)), float64(1.0-2.0*(qxx+qzz)))) * RAD_TO_DEG

	return pitch, yaw, roll
}

func MultiplyQuaternion(a, b Quaternion) Quaternion {
	return Quaternion{
		X: a.X*b.W + a.Y*b.Z - a.Z*b.Y + a.W*b.X,
		Y: -a.X*b.Z + a.Y*b.W + a.Z*b.X + a.W*b.Y,
		Z: a.X*b.Y - a.Y*b.X + a.Z*b.W + a.W*b.Z,
		W: -a.X*b.X - a.Y*b.Y - a.Z*b.Z + a.W*b.W,
	}
}

func NormalizeQuaternion(q *Quaternion) {
	mag := float32(math.Sqrt(float64(q.X*q.X + q.Y*q.Y + q.Z*q.Z + q.W*q.W)))
	if mag > 0 {
		q.X /= mag
		q.Y /= mag
		q.Z /= mag
		q.W /= mag
	}
}

func Normalize180(angle float32) float32 {
	angle = float32(math.Mod(float64(angle+180.0), 360.0))
	if angle < 0 {
		angle += 360.0
	}
	return angle - 180.0
}
