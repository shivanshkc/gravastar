package main

import (
	"fmt"
	"image/color"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shivanshkc/gravastar/pkg/physics"
)

const (
	screenWidth  = 1000
	screenHeight = 1000
	dotRadius    = 3
	dotMass      = 1
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Gravastar")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(ebiten.SyncWithFPS)

	game := NewGame()

	fmt.Println("Starting game...")
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}

type Game struct {
	engine       physics.GravityEngine
	lastTickTime time.Time
}

func NewGame() *Game {
	return &Game{
		engine:       physics.NewGravityEngine(screenWidth, screenHeight),
		lastTickTime: time.Now(),
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Update() error {
	// Handle mouse clicks to add new dots.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.addDotOnClick()
	}

	// Handle R key to reset simulation.
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.engine = physics.NewGravityEngine(screenWidth, screenHeight)
	}

	timeNow := time.Now()
	g.engine.Tick(timeNow.Sub(g.lastTickTime))
	g.lastTickTime = timeNow
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Black background.
	screen.Fill(color.Black)

	dots := g.engine.Read()
	for _, dot := range dots {
		// Convert 3D position to 2D screen coordinates.
		x, y := float32(dot.Position.X), float32(dot.Position.Y)

		// Convert color from Vec3 (0-1) to RGBA (0-255)
		dotColor := color.RGBA{
			R: uint8(dot.Color.X * 255),
			G: uint8(dot.Color.Y * 255),
			B: uint8(dot.Color.Z * 255),
			A: 255,
		}

		vector.FillCircle(screen, x, y, dotRadius, dotColor, true)
	}

	// Draw instructions.
	instructions := []string{"Left Click: Add dot", "R: Reset simulation", "Dots: " + strconv.Itoa(len(dots))}
	for i, instruction := range instructions {
		ebitenutil.DebugPrintAt(screen, instruction, 10, 10+i*15)
	}
}

func (g *Game) addDotOnClick() {
	x, y := ebiten.CursorPosition()

	dot := physics.Dot{
		ID:       uuid.NewString(),
		Mass:     dotMass,
		Radius:   dotRadius,
		Position: physics.NewVec3(float64(x), float64(y), 0),
		Velocity: physics.Vec3{},
		Color:    physics.NewRandVec3(),
	}

	_ = g.engine.AddDot(dot)
}
