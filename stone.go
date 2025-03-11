package main

import (
	"fmt"
	"math"
	"math/rand"

	r "github.com/gen2brain/raylib-go/raylib"
)

// Stone represents a stone object in the game
type Stone struct {
	X          float32
	Y          float32
	Width      int32
	Height     int32
	Texture    r.Texture2D
	Health     int32
	FlashTimer float32
	WasHit     bool
	IsGolden   bool
}

// NewStone creates a new stone instance
func NewStone(gameWidth, gameHeight int32) *Stone {
	stone := &Stone{
		Width:      32,
		Height:     32,
		Health:     rand.Int31n(10) + 3, // Stones are a bit weaker than trees
		FlashTimer: 0,
		IsGolden:   rand.Float32() < 0.2, // 20% chance to be golden
	}

	// Random position within game bounds
	stone.X = float32(rand.Float64() * float64(gameWidth-stone.Width))
	stone.Y = float32(rand.Float64() * float64(gameHeight-stone.Height))

	// Load appropriate texture
	if stone.IsGolden {
		stone.Texture = r.LoadTexture("assets/gold-stone.png")
	} else {
		stone.Texture = r.LoadTexture("assets/stone.png")
	}

	return stone
}

// OnClick handles mouse click interactions with the stone
func (s *Stone) OnClick(mouseWorldPos r.Vector2, harvestDamage int32) bool {
	if r.CheckCollisionPointRec(
		mouseWorldPos,
		r.Rectangle{
			X:      s.X,
			Y:      s.Y,
			Width:  float32(s.Width),
			Height: float32(s.Height),
		},
	) {
		s.Health -= harvestDamage
		s.FlashTimer = 0.1
		s.WasHit = true
		return s.Health <= 0
	}
	s.WasHit = false
	return false
}

// Draw renders the stone
func (s *Stone) Draw(debug bool) {
	// Draw normal sprite
	r.DrawTextureEx(
		s.Texture,
		r.Vector2{X: s.X, Y: s.Y},
		0,
		1,
		r.White,
	)

	// Draw white rectangle overlay when flashing
	if s.FlashTimer > 0 {
		s.FlashTimer -= r.GetFrameTime()
		r.DrawRectangle(
			int32(math.Floor(float64(s.X))),
			int32(math.Floor(float64(s.Y))),
			s.Width,
			s.Height,
			r.ColorAlpha(r.White, 0.5),
		)
	}

	if debug {
		r.DrawRectangleLines(
			int32(math.Floor(float64(s.X))),
			int32(math.Floor(float64(s.Y))),
			s.Width,
			s.Height,
			r.Red,
		)
		healthText := fmt.Sprintf("%d", s.Health)
		r.DrawText(
			healthText,
			int32(math.Floor(float64(s.X))),
			int32(math.Floor(float64(s.Y)-20)),
			20,
			r.White,
		)
	}
}

// Unload frees the texture from memory
func (s *Stone) Unload() {
	r.UnloadTexture(s.Texture)
}

// GetBounds returns the bounds of the stone
func (s *Stone) GetBounds() r.Rectangle {
	return r.Rectangle{
		X:      s.X,
		Y:      s.Y,
		Width:  float32(s.Width),
		Height: float32(s.Height),
	}
}
