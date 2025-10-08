package physics

const (
	gravitationalConstant = 1
)

// Dot represents a circular gravitational body.
type Dot struct {
	ID string

	Mass     float64
	Radius   float64
	Position Vec3
	Velocity Vec3

	Color Vec3
}
