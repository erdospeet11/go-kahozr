package main

import (
	r "github.com/gen2brain/raylib-go/raylib"
)

type Boss struct {
	X          float32
	Y          float32
	Width      int32
	Height     int32
	Texture    r.Texture2D
	Speed      float32
	Health     int32
	Target     *Player
	DropChance float32
	FlashTimer float32
	WasHit     bool
}

func NewBoss(x, y float32, target *Player) *Boss {
	return &Boss{
		X:          x,
		Y:          y,
		Width:      46,
		Height:     66,
		Texture:    r.LoadTexture("assets/boss.png"),
		Speed:      1.0,
		Health:     100, // More health than regular enemy
		Target:     target,
		DropChance: 1.0, // Always drops item
		FlashTimer: 0,
	}
}

func (b *Boss) Update() {
	if b.Target == nil {
		return
	}

	// Move towards player
	dx := b.Target.X - b.X
	dy := b.Target.Y - b.Y
	dist := float32(r.Vector2Length(r.Vector2{X: dx, Y: dy}))

	if dist > 0 {
		b.X += (dx / dist) * b.Speed
		b.Y += (dy / dist) * b.Speed
	}

	// Update flash effect
	if b.FlashTimer > 0 {
		b.FlashTimer -= r.GetFrameTime()
	}
}

func (b *Boss) Draw(debug bool) {
	r.DrawTextureEx(
		b.Texture,
		r.Vector2{X: b.X, Y: b.Y},
		0,
		1,
		r.White,
	)

	if b.FlashTimer > 0 {
		r.DrawRectangle(
			int32(b.X),
			int32(b.Y),
			b.Width,
			b.Height,
			r.ColorAlpha(r.White, 0.5),
		)
	}

	if debug {
		r.DrawRectangleLines(
			int32(b.X),
			int32(b.Y),
			b.Width,
			b.Height,
			r.Red,
		)
		r.DrawText(
			string(rune(b.Health+'0')),
			int32(b.X),
			int32(b.Y-20),
			20,
			r.White,
		)
	}
}

func (b *Boss) CheckCollision(player *Player) bool {
	return r.CheckCollisionRecs(
		r.Rectangle{X: b.X, Y: b.Y, Width: float32(b.Width), Height: float32(b.Height)},
		r.Rectangle{X: player.X, Y: player.Y, Width: float32(player.Width), Height: float32(player.Height)},
	)
}

func (b *Boss) TakeDamage(amount int32) {
	b.Health -= amount
	b.FlashTimer = 0.1
	b.WasHit = true
}

func (b *Boss) IsDead() bool {
	return b.Health <= 0
}

func (b *Boss) GetDropPosition() r.Vector2 {
	return r.Vector2{
		X: b.X + float32(b.Width)/2,
		Y: b.Y + float32(b.Height)/2,
	}
}

func (b *Boss) Unload() {
	r.UnloadTexture(b.Texture)
}
