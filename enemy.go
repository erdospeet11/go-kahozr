package main

import (
	"fmt"
	"math"

	r "github.com/gen2brain/raylib-go/raylib"
)

// Vector2 represents a 2D vector
type Vector2 struct {
	X float32
	Y float32
}

// Enemy represents an enemy entity in the game
type Enemy struct {
	X             float32
	Y             float32
	Width         int32
	Height        int32
	Speed         float32
	Texture       r.Texture2D
	Scale         float32
	Player        *Player
	Direction     Vector2
	FacingLeft    bool
	MaxHealth     int32
	CurrentHealth int32
	DamageText    struct {
		Value    int32
		Position r.Vector2
		Alpha    float32
		Timer    float32
	}
	DropChance     float32 // Chance to drop goodie bag (0.0 to 1.0)
	FlashTimer     float32
	DamageCooldown float32 // Add this field for rate-limiting damage
}

// NewEnemy creates a new enemy instance
func NewEnemy(x, y float32, target *Player) *Enemy {
	return &Enemy{
		X:              x,
		Y:              y,
		Width:          16,
		Height:         32,
		Texture:        r.LoadTexture("assets/enemy.png"),
		Speed:          50.0,
		Player:         target,
		MaxHealth:      3,
		CurrentHealth:  3,
		DropChance:     0.4,
		Scale:          1.0,
		DamageCooldown: 0,
	}
}

// Update updates the enemy's position and behavior
func (e *Enemy) Update() {
	deltaTime := r.GetFrameTime()

	// Update damage cooldown
	if e.DamageCooldown > 0 {
		e.DamageCooldown -= deltaTime
	}

	// Calculate direction vector towards player
	e.Direction = Vector2{
		X: e.Player.X - e.X,
		Y: e.Player.Y - e.Y,
	}

	// Update facing direction based on movement
	if e.Direction.X != 0 {
		e.FacingLeft = e.Direction.X < 0
	}

	// Normalize direction vector
	length := float32(Sqrt(float64(e.Direction.X*e.Direction.X + e.Direction.Y*e.Direction.Y)))
	if length > 0 {
		e.Direction.X /= length
		e.Direction.Y /= length
	}

	// Update position
	e.X += e.Direction.X * e.Speed * deltaTime
	e.Y += e.Direction.Y * e.Speed * deltaTime

	// Update damage text
	if e.DamageText.Timer > 0 {
		e.DamageText.Timer -= r.GetFrameTime()
		e.DamageText.Position.Y -= 1
		e.DamageText.Alpha = e.DamageText.Timer
	}
}

// Draw renders the enemy
func (e *Enemy) Draw(debug bool) {
	// Calculate sprite dimensions
	spriteWidth := float32(e.Texture.Width)
	spriteHeight := float32(e.Texture.Height)

	// Set up source rectangle with flipping
	srcRec := r.Rectangle{
		X:      0,
		Y:      0,
		Width:  spriteWidth * float32(map[bool]int{true: 1, false: -1}[e.FacingLeft]),
		Height: spriteHeight,
	}

	// Set up destination rectangle
	destRec := r.Rectangle{
		X:      e.X,
		Y:      e.Y,
		Width:  spriteWidth * e.Scale,
		Height: spriteHeight * e.Scale,
	}

	// Draw enemy
	r.DrawTexturePro(
		e.Texture,
		srcRec,
		destRec,
		r.Vector2{X: 0, Y: 0},
		0,
		r.White,
	)

	// Debug collision box
	if debug {
		r.DrawRectangleLines(
			int32(e.X),
			int32(e.Y),
			e.Width,
			e.Height,
			r.Green,
		)
	}

	// Draw health bar
	barWidth := float32(e.Width)
	barHeight := float32(4)
	healthPercent := float32(e.CurrentHealth) / float32(e.MaxHealth)
	r.DrawRectangleV(
		r.Vector2{X: e.X, Y: e.Y - barHeight - 2},
		r.Vector2{X: barWidth * healthPercent, Y: barHeight},
		r.Red,
	)

	// Draw damage text
	if e.DamageText.Timer > 0 {
		text := fmt.Sprintf("-%d", e.DamageText.Value)
		r.DrawText(
			text,
			int32(e.DamageText.Position.X),
			int32(e.DamageText.Position.Y),
			14,
			r.ColorAlpha(r.Red, e.DamageText.Alpha),
		)

		// Update position and alpha
		e.DamageText.Position.Y -= 1
		e.DamageText.Timer -= r.GetFrameTime()
		e.DamageText.Alpha = e.DamageText.Timer
	}
}

// Unload frees the texture from memory
func (e *Enemy) Unload() {
	r.UnloadTexture(e.Texture)
}

// GetBounds returns the enemy's bounding rectangle
func (e *Enemy) GetBounds() r.Rectangle {
	return r.Rectangle{
		X:      e.X,
		Y:      e.Y,
		Width:  float32(e.Width),
		Height: float32(e.Height),
	}
}

// CheckCollision checks if the enemy collides with the player
func (e *Enemy) CheckCollision(player *Player) bool {
	return r.CheckCollisionRecs(e.GetBounds(), player.GetBounds())
}

// Sqrt is a helper function for square root calculation
func Sqrt(x float64) float64 {
	return math.Sqrt(x)
}

// TakeDamage is a method to handle enemy damage
func (e *Enemy) TakeDamage(damage int32) {
	e.CurrentHealth -= damage
	if e.CurrentHealth < 0 {
		e.CurrentHealth = 0
	}

	// Update damage text
	e.DamageText = struct {
		Value    int32
		Position r.Vector2
		Alpha    float32
		Timer    float32
	}{
		Value: damage,
		Position: r.Vector2{
			X: e.X + float32(e.Width)/2 - 8,
			Y: e.Y - 10,
		},
		Alpha: 1.0,
		Timer: 1.0,
	}
}

// Add IsDead method
func (e *Enemy) IsDead() bool {
	return e.CurrentHealth <= 0
}

// Add method to handle drops
func (e *Enemy) GetDropPosition() r.Vector2 {
	return r.Vector2{
		X: e.X + float32(e.Width)/2,
		Y: e.Y + float32(e.Height)/2,
	}
}
