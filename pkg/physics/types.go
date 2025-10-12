package physics

const (
	gravitationalConstant = 1
)

// Dot represents a circular gravitational body.
type Dot struct {
	ID string `json:"id"`

	Mass     float64 `json:"mass"`
	Radius   float64 `json:"radius"`
	Position Vec3    `json:"position"`
	Velocity Vec3    `json:"velocity"`

	Color Vec3 `json:"color"`
}
