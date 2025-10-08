package physics

import (
	"maps"
	"runtime"
	"sync"
	"time"
)

// GravityEngine encapsulates methods to manage a gravity simulation.
type GravityEngine interface {
	// Tick advances the simulation further in time.
	Tick()

	// Read the current state of the simulation.
	Read() map[string]Dot

	// AddDot adds a new Dot to the simulation.
	AddDot(Dot)

	// RemoveDot removes the Dot with the given ID from the simulation.
	//
	// If no Dot is found with the given ID, the returned param is false, otherwise true.
	RemoveDot(id string) bool
}

// gravityEngine implements the GravityEngine interface.
type gravityEngine struct {
	dots   map[string]Dot
	mutex  *sync.RWMutex
	width  float64
	height float64
}

// Tick calculates the gravitational effect on each Dot due to all other Dots and updates its
// properties (position, velocity etc). Since these calculations are time-dependent, Tick uses
// 10 milliseconds as the time delta.
//
// For efficiency, Tick uses goroutines for calculation. The number of goroutines launched is
// equal to runtime.NumCPU().
func (g *gravityEngine) Tick() {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Delta time. This is required for motion physics.
	delta := (10 * time.Millisecond).Seconds()
	// If no dots, nothing to calculate.
	if len(g.dots) == 0 {
		return
	}

	// To control the batch size.
	semaphore := make(chan struct{}, runtime.NumCPU())
	defer close(semaphore)

	// To receive calculation results.
	calculations := make(chan Dot, len(g.dots))
	defer close(calculations)

	// Create a snapshot of dots to avoid concurrent map access
	dotsClone := maps.Clone(g.dots)

	// Loop to calculate the new properties for each dot.
	for dotID, dot := range dotsClone {
		dotID, dot := dotID, dot

		// Obtain a spot in the batch.
		semaphore <- struct{}{}
		go func() {
			// Release the spot.
			defer func() { <-semaphore }()

			// Vector to aggregate all the accelerations.
			totalAcceleration := Vec3{}

			// Loop over all other dots to calculate their effect on this one.
			for otherDotID, otherDot := range dotsClone {
				if otherDotID == dotID {
					continue
				}

				// Distance calculation.
				distance := otherDot.Position.Sub(dot.Position)
				// Scale the distance down to avoid crunching unnecessarily large numbers.
				distance = distance.Div(1000)

				// Denominator calculation as per the r-squared rule.
				distanceMag := distance.Mag()
				distanceMagSquared := distanceMag * distanceMag

				// Softening to avoid singularities.
				softeningSquared := 0.05 * 0.05
				// The last distanceMag multiplication is due to vector form of the gravity equation,
				// and so, softening should not be added to it.
				denominator := (distanceMagSquared + softeningSquared) * distanceMag

				// Acceleration calculation.
				acceleration := distance.Mul(gravitationalConstant * otherDot.Mass / denominator)
				totalAcceleration = totalAcceleration.Add(acceleration)
			}

			// Second law of motion to calculate the displacement due to acceleration.
			halfAtSquared := totalAcceleration.Mul(0.5 * delta * delta)
			displacement := dot.Velocity.Mul(delta).Add(halfAtSquared)
			dot.Position = dot.Position.Add(displacement)

			// First law of motion to calculate the final velocity of the dot.
			dot.Velocity = dot.Velocity.Add(totalAcceleration.Mul(delta))

			// Wall collision detection and response
			// Left wall collision
			if dot.Position.X-dot.Radius <= 0 {
				dot.Position.X = dot.Radius
				dot.Velocity.X = -dot.Velocity.X
			}
			// Right wall collision
			if dot.Position.X+dot.Radius >= g.width {
				dot.Position.X = g.width - dot.Radius
				dot.Velocity.X = -dot.Velocity.X
			}
			// Top wall collision
			if dot.Position.Y-dot.Radius <= 0 {
				dot.Position.Y = dot.Radius
				dot.Velocity.Y = -dot.Velocity.Y
			}
			// Bottom wall collision
			if dot.Position.Y+dot.Radius >= g.height {
				dot.Position.Y = g.height - dot.Radius
				dot.Velocity.Y = -dot.Velocity.Y
			}

			calculations <- dot
		}()
	}

	// Receive calculation results.
	for range dotsClone {
		dot := <-calculations
		g.dots[dot.ID] = dot
	}
}

func (g *gravityEngine) Read() map[string]Dot {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	return maps.Clone(g.dots)
}

func (g *gravityEngine) AddDot(dot Dot) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.dots[dot.ID] = dot
}

func (g *gravityEngine) RemoveDot(id string) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, exists := g.dots[id]; !exists {
		return false
	}

	delete(g.dots, id)
	return true
}

// NewGravityEngine returns a new GravityEngine implementation.
func NewGravityEngine(width, height float64) GravityEngine {
	return &gravityEngine{
		dots:   make(map[string]Dot),
		mutex:  &sync.RWMutex{},
		width:  width,
		height: height,
	}
}
