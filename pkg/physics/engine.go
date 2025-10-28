package physics

import (
	"context"
	"errors"
	"maps"
	"sync"
	"time"
)

// ErrDotAlreadyExists is returned when a new dot is attempted to be added to the Engine but its ID is not unique.
var ErrDotAlreadyExists = errors.New("dot with same ID already exists")

// ErrEngineClosed is returned when operations are attempted on a closed engine.
var ErrEngineClosed = errors.New("engine is closed")

// GravityEngine encapsulates methods to manage a gravity simulation.
type GravityEngine struct {
	dots   map[string]Dot
	mutex  *sync.RWMutex
	width  int
	height int

	collisionChan chan Dot
	closedChan    chan struct{}
}

// Size returns the width and height of the simulation respectively.
func (g *GravityEngine) Size() (int, int) {
	return g.width, g.height
}

// Collisions returns a read-only channel that emits dots that collided with one of the four walls.
func (g *GravityEngine) Collisions() <-chan Dot {
	return g.collisionChan
}

// Read returns a direct reference to the internal dots map for efficient access.
//
// WARNING: The returned map is read-only and must NOT be modified. Modifying the map
// will cause undefined behavior and potential data races.
//
// The map's content may change between calls as the simulation progresses.
// If you need a stable snapshot that won't change, use ReadSnapshot() instead.
func (g *GravityEngine) Read() map[string]Dot {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.dots
}

// ReadSnapshot returns an independent copy of the current simulation state.
//
// Unlike Read(), this returns a snapshot that will not be affected by subsequent
// simulation ticks. Use this when you need to store or process the state over time.
//
// Warning: This method clones the entire dots map, which is expensive for large
// simulations. Prefer Read() if you can tolerate direct references and don't need
// a stable snapshot.
func (g *GravityEngine) ReadSnapshot() map[string]Dot {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return maps.Clone(g.dots)
}

// AddDot adds a new Dot to the simulation. It requires the dot to have a unique ID.
// If it's not unique, ErrDotAlreadyExists is returned.
// Returns ErrEngineClosed if the engine has been closed.
func (g *GravityEngine) AddDot(dot Dot) error {
	// Check if engine is closed by trying to receive from closedChan (non-blocking)
	select {
	case <-g.closedChan:
		return ErrEngineClosed
	default:
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, ok := g.dots[dot.ID]; ok {
		return ErrDotAlreadyExists
	}

	g.dots[dot.ID] = dot
	return nil
}

// RemoveDot removes the Dot with the given ID from the simulation.
//
// If no Dot is found with the given ID, the returned param is false, otherwise true.
// Returns false if the engine has been closed.
func (g *GravityEngine) RemoveDot(id string) bool {
	// Check if engine is closed
	select {
	case <-g.closedChan:
		return false
	default:
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()

	if _, exists := g.dots[id]; !exists {
		return false
	}

	delete(g.dots, id)
	return true
}

// RemoveAll clears the simulation/engine.
func (g *GravityEngine) RemoveAll() {
	// Check if engine is closed
	select {
	case <-g.closedChan:
		return
	default:
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()

	g.dots = map[string]Dot{}
}

// Run the simulation. This method returns when either the context expires or Close() is called.
//
// Once Run returns, call Close() to clean up resources if it hasn't been called already.
func (g *GravityEngine) Run(ctx context.Context, targetFPS uint) {
	// Ticker provides an efficient way to run the simulation at the given target FPS.
	ticker := time.NewTicker(time.Second / time.Duration(targetFPS))
	defer ticker.Stop()

	// This will be used to calculate delta for the Tick method call.
	timeLast := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-g.closedChan:
			return
		case <-ticker.C:
			timeNow := time.Now()
			g.Tick(timeNow.Sub(timeLast))
			timeLast = timeNow
		}
	}
}

// Tick advances the simulation by one time step. It calculates the gravitational effect on each Dot due to all other
// Dots and updates its properties (position, velocity etc.). This method makes the simulation compatible with game
// engines like Ebiten.
//
// For efficiency, Tick uses goroutines for calculation. The number of goroutines launched is equal to the number of dots.
//
// Tick can be called manually for integration with external game loops. Do not call Tick while Run is active -
// they share the same mutex and will result in unpredictable timing.
func (g *GravityEngine) Tick(delta time.Duration) {
	// Check if engine is closed
	select {
	case <-g.closedChan:
		return
	default:
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Perform the physics simulation.
	tick(g.width, g.height, g.dots, delta, g.collisionChan)
}

// Close shuts down the engine and releases resources.
//
// This closes the collision channel, which will cause any readers to eventually
// receive a closed channel signal. If Run() is active, it will return immediately.
// Once closed, the engine cannot be reused.
//
// It's safe to call Close multiple times.
func (g *GravityEngine) Close() error {
	// Check if already closed (non-blocking check)
	select {
	case <-g.closedChan:
		return nil
	default:
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Close both channels to signal shutdown
	close(g.collisionChan)
	close(g.closedChan)
	g.dots = nil
	return nil
}

// NewGravityEngine returns a new GravityEngine implementation.
func NewGravityEngine(width, height int) *GravityEngine {
	return &GravityEngine{
		dots:          make(map[string]Dot),
		mutex:         &sync.RWMutex{},
		width:         width,
		height:        height,
		collisionChan: make(chan Dot),
		closedChan:    make(chan struct{}),
	}
}

// RightWallCollisionRemover starts an infinite loop to listen on the Collisions channel of the engine.
//
// Any dot colliding with the right wall is deleted from the engine.
//
// The context parameter can be used to terminate the infinite loop.
//
// Note: This function spawns a goroutine for each removal to avoid deadlock with Tick's mutex.
func RightWallCollisionRemover(ctx context.Context, engine *GravityEngine) {
	collisionChan := engine.Collisions()
	width, _ := engine.Size()

	for {
		select {
		case <-ctx.Done():
			return
		case dot, ok := <-collisionChan:
			if !ok {
				// Channel closed, engine is shutting down
				return
			}

			if dot.Position.X+dot.Radius >= float64(width) {
				// Launch in goroutine to avoid deadlock.
				go engine.RemoveDot(dot.ID)
			}
		}
	}
}
