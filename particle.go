package main

import (
	"math"
	"math/rand"

	r "github.com/gen2brain/raylib-go/raylib"
)

type Particle struct {
	Position r.Vector2
	Velocity r.Vector2
	Color    r.Color
	Life     float32
	Size     float32
	Text     string
}

type ParticleSystem struct {
	Particles []Particle
}

func NewParticleSystem() *ParticleSystem {
	return &ParticleSystem{
		Particles: make([]Particle, 0),
	}
}

func (ps *ParticleSystem) SpawnExplosion(color r.Color, particleCount int, x, y float32) {
	for i := 0; i < particleCount; i++ {
		// Random angle and speed for each particle
		angle := rand.Float32() * 2 * 3.14159 // Random angle in radians
		speed := 1.0 + rand.Float32()*2.0     // Random speed between 2 and 5

		particle := Particle{
			Position: r.Vector2{X: x, Y: y},
			Velocity: r.Vector2{
				X: float32(speed * float32(math.Cos(float64(angle)))),
				Y: float32(speed * float32(math.Sin(float64(angle)))),
			},
			Color: color,
			Life:  0.25, // Life in seconds
			Size:  4.0,
		}
		ps.Particles = append(ps.Particles, particle)
	}
}

func (ps *ParticleSystem) Update() {
	deltaTime := r.GetFrameTime()
	var activeParticles []Particle

	for _, p := range ps.Particles {
		p.Life -= deltaTime
		if p.Life > 0 {
			// Update position
			p.Position.X += p.Velocity.X
			p.Position.Y += p.Velocity.Y

			// Fade out
			//p.Color.A = uint8(255 * (p.Life))
			p.Size *= 0.99

			activeParticles = append(activeParticles, p)
		}
	}

	ps.Particles = activeParticles
}

func (ps *ParticleSystem) Draw() {
	for _, p := range ps.Particles {
		if p.Text != "" {
			// Draw text particles (damage numbers)
			alpha := uint8(255.0 * (p.Life / 1.0))
			color := r.Color{R: p.Color.R, G: p.Color.G, B: p.Color.B, A: alpha}
			r.DrawText(p.Text, int32(p.Position.X), int32(p.Position.Y), int32(p.Size), color)
		} else {
			// Draw regular particles
			r.DrawRectangleV(p.Position, r.Vector2{X: p.Size, Y: p.Size}, p.Color)
		}
	}
}

func (ps *ParticleSystem) SpawnDamageNumber(text string, x, y float32) {
	const floatSpeed float32 = -1.5 // Reduced from -30.0 to -1.5
	const lifetime float32 = 1.0    // Keep the same lifetime

	particle := Particle{
		Position: r.Vector2{X: x, Y: y},
		Velocity: r.Vector2{X: 0, Y: floatSpeed},
		Color:    r.White,
		Life:     lifetime,
		Size:     20.0,
		Text:     text,
	}
	ps.Particles = append(ps.Particles, particle)
}
