package physics

import (
	"sync"
	"time"
)

// tick is a pure function that calculates gravitational interaction between the given dots. It also detects dots
// colliding with walls as per the given width and height. These collisions are reported to the collisionChan.
func tick(width, height int, dots map[string]Dot, delta time.Duration, collisionChan chan Dot) {
	// If no dots, nothing to calculate.
	if len(dots) == 0 {
		return
	}

	// Convert delta to seconds for SI units.
	deltaSec := delta.Seconds()

	// WaitGroup to ensure all goroutines complete before updating the map.
	var wg sync.WaitGroup

	// Buffered channel to receive calculation results without blocking goroutines.
	// Buffer size matches the number of dots to prevent any blocking on sends.
	calculations := make(chan Dot, len(dots))

	// Spawn one goroutine per dot to calculate its new state.
	wg.Add(len(dots))

	for dotID, dot := range dots {
		dotID, dot := dotID, dot

		go func() {
			defer wg.Done()

			// Vector to aggregate all gravitational accelerations acting on this dot.
			totalAcceleration := Vec3{}

			// Calculate gravitational force from every other dot.
			for otherDotID, otherDot := range dots {
				// Skip self-interaction.
				if otherDotID == dotID {
					continue
				}

				// Calculate distance vector from this dot to the other dot.
				distance := otherDot.Position.Sub(dot.Position)

				// Scale distance down by 1000 to avoid working with very large numbers.
				// This is a numerical stability optimization - seems to work well with G = 1.
				distance = distance.Div(1000)

				// Calculate distance magnitude and its square for the gravity equation.
				distanceMag := distance.Mag()
				distanceMagSquared := distanceMag * distanceMag

				// Softening factor to prevent singularities when dots get very close.
				// Without this, acceleration would approach infinity as distance approaches zero.
				softeningSquared := 0.05 * 0.05

				// Denominator for gravity equation: (r² + ε²) * r
				// The extra distanceMag factor is for the vector form of the equation.
				denominator := (distanceMagSquared + softeningSquared) * distanceMag

				// Calculate acceleration using Newton's law of gravitation: a = G * m / r²
				// Vector form ensures acceleration points toward the other dot.
				acceleration := distance.Mul(gravitationalConstant * otherDot.Mass / denominator)
				totalAcceleration = totalAcceleration.Add(acceleration)
			}

			// Second law of motion to calculate displacement.
			halfAtSquared := totalAcceleration.Mul(0.5 * deltaSec * deltaSec)
			displacement := dot.Velocity.Mul(deltaSec).Add(halfAtSquared)
			dot.Position = dot.Position.Add(displacement)

			// First law of motion to calculate final velocity.
			dot.Velocity = dot.Velocity.Add(totalAcceleration.Mul(deltaSec))

			// Track whether any wall collision occurred.
			var collisionOccurred bool

			// Check and respond to collisions with the four walls.
			// Use perfect elastic collisions (velocity reversal).

			// Left wall collision.
			if dot.Position.X-dot.Radius <= 0 {
				dot.Position.X = dot.Radius // Prevent dot from going through wall.
				dot.Velocity.X = -dot.Velocity.X
				collisionOccurred = true
			}

			// Right wall collision.
			if dot.Position.X+dot.Radius >= float64(width) {
				dot.Position.X = float64(width) - dot.Radius
				dot.Velocity.X = -dot.Velocity.X
				collisionOccurred = true
			}

			// Top wall collision.
			if dot.Position.Y-dot.Radius <= 0 {
				dot.Position.Y = dot.Radius
				dot.Velocity.Y = -dot.Velocity.Y
				collisionOccurred = true
			}

			// Bottom wall collision.
			if dot.Position.Y+dot.Radius >= float64(height) {
				dot.Position.Y = float64(height) - dot.Radius
				dot.Velocity.Y = -dot.Velocity.Y
				collisionOccurred = true
			}

			// Report collision if one occurred.
			// This is a blocking send - if the reader isn't keeping up, the simulation will block.
			// This is intentional to expose backpressure issues during development.
			if collisionOccurred {
				collisionChan <- dot
			}

			// Send the updated dot to the results channel.
			calculations <- dot
		}()
	}

	// Wait for all goroutines to finish their calculations.
	wg.Wait()

	// Close the calculations channel since no more results will be sent.
	close(calculations)

	// Collect all calculation results and update the dots map.
	// This happens after all reads are complete, so it's safe to write to the map.
	for dot := range calculations {
		dots[dot.ID] = dot
	}
}
