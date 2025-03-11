package main

import (
	"math/rand"

	r "github.com/gen2brain/raylib-go/raylib"
)

type Portal struct {
	X          float32
	Y          float32
	Width      int32
	Height     int32
	Texture    r.Texture2D
	SpawnTimer float32
	SpawnRate  float32
	SpawnCount int
	MaxSpawns  int
	IsDone     bool
}

func NewPortal(gameWidth, gameHeight int32, game *Game) *Portal {
	const padding float32 = 100
	var x, y float32
	var bounds r.Rectangle

	// Keep trying positions until we find an unoccupied spot
	for {
		x = padding + float32(rand.Float64()*float64(float32(gameWidth)-2*padding))
		y = padding + float32(rand.Float64()*float64(float32(gameHeight)-2*padding))
		bounds = r.Rectangle{
			X:      x,
			Y:      y,
			Width:  16,
			Height: 32,
		}
		if !game.IsPositionOccupied(bounds, 30) {
			break
		}
	}

	return &Portal{
		X:          x,
		Y:          y,
		Width:      16,
		Height:     32,
		Texture:    r.LoadTexture("assets/portal.png"),
		SpawnRate:  5.0,
		SpawnTimer: 5.0,
		SpawnCount: 0,
		MaxSpawns:  10,
		IsDone:     false,
	}
}

func (p *Portal) Update(deltaTime float32) bool {
	if p.SpawnCount >= p.MaxSpawns {
		p.IsDone = true
		return false
	}

	p.SpawnTimer -= deltaTime
	if p.SpawnTimer <= 0 {
		p.SpawnTimer = p.SpawnRate
		p.SpawnCount++
		return true
	}
	return false
}

func (p *Portal) Draw(debug bool) {
	r.DrawTextureEx(
		p.Texture,
		r.Vector2{X: p.X, Y: p.Y},
		0,
		1,
		r.White,
	)

	if debug {
		r.DrawRectangleLines(
			int32(p.X),
			int32(p.Y),
			p.Width,
			p.Height,
			r.Purple,
		)
	}
}

func (p *Portal) Unload() {
	r.UnloadTexture(p.Texture)
}

// Add method to get spawn position
func (p *Portal) GetSpawnPosition() r.Vector2 {
	// Spawn enemy slightly offset from portal center
	offset := float32(rand.Float32()*40 - 20) // Random offset between -20 and 20
	return r.Vector2{
		X: p.X + float32(p.Width)/2 + offset,
		Y: p.Y + float32(p.Height)/2 + offset,
	}
}
