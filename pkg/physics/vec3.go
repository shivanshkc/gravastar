package physics

import (
	"math"
	"math/rand"
	"time"
)

// Vec3 is a 3D vector.
type Vec3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// NewVec3 creates a new Vec3.
func NewVec3(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

// NewRandVec3 returns a new random Vec3 with each dimension in [0, 1).
func NewRandVec3() Vec3 {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	return NewVec3(random.Float64(), random.Float64(), random.Float64())
}

// Add adds the given vector to this vector
// and returns the result.
func (v Vec3) Add(arg Vec3) Vec3 {
	return NewVec3(v.X+arg.X, v.Y+arg.Y, v.Z+arg.Z)
}

// Sub subtracts the given vector from this vector
// and returns the result.
func (v Vec3) Sub(arg Vec3) Vec3 {
	return NewVec3(v.X-arg.X, v.Y-arg.Y, v.Z-arg.Z)
}

// Mul multiplies the vector with the given argument
// and returns the result.
func (v Vec3) Mul(arg float64) Vec3 {
	return NewVec3(v.X*arg, v.Y*arg, v.Z*arg)
}

// Div divides the vector with the given argument
// and returns the result.
func (v Vec3) Div(arg float64) Vec3 {
	return NewVec3(v.X/arg, v.Y/arg, v.Z/arg)
}

// Dot calculates the dot product of this vector with the given vector.
func (v Vec3) Dot(arg Vec3) float64 {
	return v.X*arg.X + v.Y*arg.Y + v.Z*arg.Z
}

// DotSelf returns the dot product of the vector with itself.
// It is equivalent to its magnitude squared.
func (v Vec3) DotSelf() float64 {
	return v.Dot(v)
}

// Cross calculates the cross product of this vector with the given vector
// and returns the result.
func (v Vec3) Cross(arg Vec3) Vec3 {
	return NewVec3(
		v.Y*arg.Z-v.Z*arg.Y,
		v.Z*arg.X-v.X*arg.Z,
		v.X*arg.Y-v.Y*arg.X,
	)
}

// Mag calculates the magnitude of the vector.
func (v Vec3) Mag() float64 {
	return math.Sqrt(v.DotSelf())
}

// Dir calculates the direction (or unit vector) of this vector.
func (v Vec3) Dir() Vec3 {
	return v.Div(v.Mag())
}

// Reflected calculates and returns the reflection of this vector
// for the given normal.
//
// To understand the formula, go to -
// https://raytracing.github.io/books/RayTracingInOneWeekend.html#metal/mirroredlightreflection
func (v Vec3) Reflected(normal Vec3) Vec3 {
	return v.Sub(normal.Mul(v.Dot(normal) * 2))
}

// IsNearZero returns true if ALL components of the vector are "very" close to zero.
func (v Vec3) IsNearZero() bool {
	precision := 0.00001
	return v.X < precision && v.Y < precision && v.Z < precision
}
