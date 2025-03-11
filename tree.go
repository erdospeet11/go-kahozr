package main

import (
	"fmt"
	"math"
	"math/rand"

	r "github.com/gen2brain/raylib-go/raylib"
)

// Tree represents a tree object in the game
type Tree struct {
	X          float32
	Y          float32
	Width      int32
	Height     int32
	Texture    r.Texture2D
	Health     int32
	FlashTimer float32
	WasHit     bool
}

// NewTree creates a new tree instance
func NewTree(gameWidth, gameHeight int32) *Tree {
	tree := &Tree{
		Width:      48,
		Height:     64,
		Health:     rand.Int31n(7) + 2,
		FlashTimer: 0,
	}

	// Random position within game bounds
	tree.X = float32(rand.Float64() * float64(gameWidth-tree.Width))
	tree.Y = float32(rand.Float64() * float64(gameHeight-tree.Height))
	tree.Texture = r.LoadTexture("assets/tree.png")

	return tree
}

// OnClick handles mouse click interactions with the tree
func (t *Tree) OnClick(mouseWorldPos r.Vector2, harvestDamage int32) bool {
	if r.CheckCollisionPointRec(
		mouseWorldPos,
		r.Rectangle{
			X:      t.X,
			Y:      t.Y,
			Width:  float32(t.Width),
			Height: float32(t.Height),
		},
	) {
		t.Health -= harvestDamage
		t.FlashTimer = 0.1
		t.WasHit = true
		return t.Health <= 0
	}
	t.WasHit = false
	return false
}

// Draw renders the tree
func (t *Tree) Draw(debug bool) {
	// Draw normal sprite
	r.DrawTextureEx(
		t.Texture,
		r.Vector2{X: t.X, Y: t.Y},
		0,
		1,
		r.White,
	)

	// Draw white rectangle overlay when flashing
	if t.FlashTimer > 0 {
		t.FlashTimer -= r.GetFrameTime()
		r.DrawRectangle(
			int32(math.Floor(float64(t.X))),
			int32(math.Floor(float64(t.Y))),
			t.Width,
			t.Height,
			r.ColorAlpha(r.White, 0.5),
		)
	}

	if debug {
		r.DrawRectangleLines(
			int32(math.Floor(float64(t.X))),
			int32(math.Floor(float64(t.Y))),
			t.Width,
			t.Height,
			r.Red,
		)
		// Use fmt.Sprintf to properly convert int to string
		healthText := fmt.Sprintf("%d", t.Health)
		r.DrawText(
			healthText,
			int32(math.Floor(float64(t.X))),
			int32(math.Floor(float64(t.Y)-20)),
			20,
			r.White,
		)
	}
}

// Unload frees the texture from memory
func (t *Tree) Unload() {
	r.UnloadTexture(t.Texture)
}

// Add this method to Tree struct
func (t *Tree) GetBounds() r.Rectangle {
	return r.Rectangle{
		X:      t.X,
		Y:      t.Y,
		Width:  float32(t.Width),
		Height: float32(t.Height),
	}
}
