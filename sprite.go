package main

import (
	"math/rand"

	r "github.com/gen2brain/raylib-go/raylib"
)

// Sprite represents a basic sprite object in the game
type Sprite struct {
	X       float32
	Y       float32
	Width   int32
	Height  int32
	Texture r.Texture2D
}

// NewSprite creates a new sprite instance
func NewSprite(imagePath string) *Sprite {
	sprite := &Sprite{
		Width:  16,
		Height: 16,
	}

	// Random position within game bounds
	sprite.X = float32(rand.Float64() * float64(GameWidth-sprite.Width))
	sprite.Y = float32(rand.Float64() * float64(GameHeight-sprite.Height))
	sprite.Texture = r.LoadTexture(imagePath)

	return sprite
}

// Draw renders the sprite
func (s *Sprite) Draw() {
	r.DrawTextureEx(
		s.Texture,
		r.Vector2{X: s.X, Y: s.Y},
		0,
		1,
		r.White,
	)
}

// Unload frees the texture from memory
func (s *Sprite) Unload() {
	r.UnloadTexture(s.Texture)
}
