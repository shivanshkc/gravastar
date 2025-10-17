package physics

import (
	"context"
	"errors"
	"maps"
	"runtime"
	"sync"
	"time"
)

// ErrDotAlreadyExists is returned when a new dot is attempted to be added to the Engine but its ID is not unique.
var ErrDotAlreadyExists = errors.New("dot with same ID already exists")

// GravityEngine encapsulates methods to manage a gravity simulation.
type GravityEngine interface {
	// Size returns the width and height of the simulation respectively.
	Size() (int, int)

	// Run the simulation. This method returns only when the context expires.
	Run(ctx context.Context, targetFPS uint)

	// Tick advances the simulation by one step. It makes GravityEngine compatible with Game Engines like Ebiten.
	//
	// Tick and Run use the same mutex under the hoods so they can be called simultaneously.
	Tick(delta time.Duration)

	// Read the current state of the simulation.
	Read() map[string]Dot

	// AddDot adds a new Dot to the simulation. It requires the dot to have a unique ID.
	// If it's not unique, ErrDotAlreadyExists is returned.
	AddDot(Dot) error

	// RemoveDot removes the Dot with the given ID from the simulation.
	//
	// If no Dot is found with the given ID, the returned param is false, otherwise true.
	RemoveDot(id string) bool
}

// gravityEngine implements the GravityEngine interface.
type gravityEngine struct {
	dots   map[string]Dot
	mutex  *sync.RWMutex
	width  int
	height int
}

func (g *gravityEngine) Size() (int, int) {
	return g.width, g.height
}

func (g *gravityEngine) Run(ctx context.Context, targetFPS uint) {
	// Ticker provides an efficient way to run the simulation at the given target FPS.
	ticker := time.NewTicker(time.Second / time.Duration(targetFPS))
	defer ticker.Stop()

	timeLast := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			timeNow := time.Now()
			// Submit precise delta for correct physics calculations.
			g.Tick(timeNow.Sub(timeLast))
			timeLast = timeNow
		}
	}
}

// Tick calculates the gravitational effect on each Dot due to all other Dots and updates its properties
// (position, velocity etc.). Since these calculations are time-dependent, Tick accepts a time delta.
//
// For efficiency, Tick uses goroutines for calculation. The number of goroutines launched is equal to runtime.NumCPU().
func (g *gravityEngine) Tick(delta time.Duration) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// If no dots, nothing to calculate.
	if len(g.dots) == 0 {
		return
	}

	// SI units.
	deltaSec := delta.Seconds()

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
			halfAtSquared := totalAcceleration.Mul(0.5 * deltaSec * deltaSec)
			displacement := dot.Velocity.Mul(deltaSec).Add(halfAtSquared)
			dot.Position = dot.Position.Add(displacement)

			// First law of motion to calculate the final velocity of the dot.
			dot.Velocity = dot.Velocity.Add(totalAcceleration.Mul(deltaSec))

			// Wall collision detection and response
			// Left wall collision
			if dot.Position.X-dot.Radius <= 0 {
				dot.Position.X = dot.Radius
				dot.Velocity.X = -dot.Velocity.X
			}
			// Right wall collision
			if dot.Position.X+dot.Radius >= float64(g.width) {
				dot.Position.X = float64(g.width) - dot.Radius
				dot.Velocity.X = -dot.Velocity.X
			}
			// Top wall collision
			if dot.Position.Y-dot.Radius <= 0 {
				dot.Position.Y = dot.Radius
				dot.Velocity.Y = -dot.Velocity.Y
			}
			// Bottom wall collision
			if dot.Position.Y+dot.Radius >= float64(g.height) {
				dot.Position.Y = float64(g.height) - dot.Radius
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

func (g *gravityEngine) AddDot(dot Dot) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, ok := g.dots[dot.ID]; ok {
		return ErrDotAlreadyExists
	}

	g.dots[dot.ID] = dot
	return nil
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
func NewGravityEngine(width, height int) GravityEngine {
	return &gravityEngine{
		dots:   make(map[string]Dot),
		mutex:  &sync.RWMutex{},
		width:  width,
		height: height,
	}
}
