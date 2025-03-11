package main

import (
	"math"

	r "github.com/gen2brain/raylib-go/raylib"
)

// DroppedItem represents an item that can be picked up in the game
type DroppedItem struct {
	X         float32
	Y         float32
	Width     int32
	Height    int32
	Texture   r.Texture2D
	Name      string
	ImagePath string
}

// NewDroppedItem creates a new dropped item instance
func NewDroppedItem(x, y float32, imagePath, name string) *DroppedItem {
	return &DroppedItem{
		X:         x,
		Y:         y,
		Width:     16,
		Height:    16,
		Texture:   r.LoadTexture(imagePath),
		Name:      name,
		ImagePath: imagePath,
	}
}

// Draw renders the dropped item
func (d *DroppedItem) Draw(debug bool) {
	r.DrawTextureEx(
		d.Texture,
		r.Vector2{X: d.X, Y: d.Y},
		0,
		1,
		r.White,
	)

	if debug {
		r.DrawRectangleLines(
			int32(math.Floor(float64(d.X))),
			int32(math.Floor(float64(d.Y))),
			d.Width,
			d.Height,
			r.Blue,
		)
	}
}

// CheckCollision checks if the item collides with the player
func (d *DroppedItem) CheckCollision(playerBounds r.Rectangle) bool {
	itemBounds := r.Rectangle{
		X:      d.X,
		Y:      d.Y,
		Width:  float32(d.Width),
		Height: float32(d.Height),
	}

	return r.CheckCollisionRecs(itemBounds, playerBounds)
}

// Unload frees the texture from memory
func (d *DroppedItem) Unload() {
	r.UnloadTexture(d.Texture)
}
