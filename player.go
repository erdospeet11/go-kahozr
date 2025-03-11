package main

import (
	r "github.com/gen2brain/raylib-go/raylib"
)

// Player represents the player entity in the game
type Player struct {
	X             float32
	Y             float32
	Width         int32
	Height        int32
	Speed         float32
	Texture       r.Texture2D
	Scale         float32
	GameWidth     int32
	GameHeight    int32
	FrameWidth    int32
	FrameHeight   int32
	CurrentFrame  int32
	FrameCount    int32
	FrameTime     float32
	IsMoving      bool
	FacingLeft    bool
	MaxHealth     int32
	CurrentHealth int32
	IsAiming      bool
	SlashAnim     struct {
		Texture     r.Texture2D
		Active      bool
		Frame       int32
		FrameTime   float32
		FrameWidth  int32
		FrameHeight int32
	}
	InvincibleTimer float32 // Time remaining for invincibility
	InvincibleTime  float32 // How long invincibility lasts

	DashSpeed         float32
	DashDuration      float32
	DashTimer         float32
	IsDashing         bool
	DashCooldown      float32
	DashCooldownTimer float32

	LastMoveDirection r.Vector2 // Track last movement direction for dash
	RegenTimer        float32
	RegenInterval     float32 // Time between each health regen tick
	CurrentWeapon     Weapon
	Weapons           []Weapon
	GhostTrail        []struct {
		Position   r.Vector2
		Alpha      float32
		FacingLeft bool
	}
	GhostTrailLength int
	Experience       int
	NextLevelExp     int

	HarvestDamage int32
}

// NewPlayer creates a new player instance
func NewPlayer(x, y float32, gameWidth, gameHeight int32) *Player {
	raygun := NewRayGun() // Create raygun first
	p := &Player{
		X:                 x,
		Y:                 y,
		Width:             16,
		Height:            16,
		Speed:             1.5,
		Scale:             1,
		Texture:           r.LoadTexture("assets/soma_fut.png"),
		GameWidth:         gameWidth,
		GameHeight:        gameHeight,
		FrameWidth:        16,
		FrameHeight:       16,
		CurrentFrame:      3,
		FrameCount:        4,
		FrameTime:         0,
		IsMoving:          false,
		FacingLeft:        false,
		MaxHealth:         100,
		CurrentHealth:     100,
		InvincibleTime:    1.0, // 1 second of invincibility after hit
		InvincibleTimer:   0,
		DashSpeed:         5.0,
		DashDuration:      0.2, // Dash lasts 0.2 seconds
		DashTimer:         0,
		IsDashing:         false,
		DashCooldown:      1.0, // 1 second cooldown
		DashCooldownTimer: 0,
		LastMoveDirection: r.Vector2{X: 1, Y: 0}, // Default right direction
		RegenTimer:        0,
		RegenInterval:     2.0,                                       // 2 seconds between each health regen
		CurrentWeapon:     raygun,                                    // Set initial weapon
		Weapons:           []Weapon{raygun, NewSword(), NewPistol()}, // Use the same raygun instance
		GhostTrail: make([]struct {
			Position   r.Vector2
			Alpha      float32
			FacingLeft bool
		}, 0),
		GhostTrailLength: 5, // Number of ghost images to show
		Experience:       0,
		NextLevelExp:     100, // Experience needed for next level
		HarvestDamage:    1,
	}

	p.SlashAnim = struct {
		Texture     r.Texture2D
		Active      bool
		Frame       int32
		FrameTime   float32
		FrameWidth  int32
		FrameHeight int32
	}{
		Texture:     r.LoadTexture("assets/slash.png"),
		FrameWidth:  65, // 390/6 frames = 65 pixels per frame
		FrameHeight: 27,
	}

	return p
}

// Update updates the player's position based on input
func (p *Player) Update(camera r.Camera2D) {
	deltaTime := r.GetFrameTime()

	// Handle weapon activation
	if r.IsMouseButtonDown(1) { // 1 is right mouse button
		p.CurrentWeapon.OnActivate(p, camera)
	} else {
		p.CurrentWeapon.OnDeactivate(p)
	}

	// Update current weapon
	p.CurrentWeapon.Update(deltaTime, p)

	// Update dash cooldown
	if p.DashCooldownTimer > 0 {
		p.DashCooldownTimer -= deltaTime
	}

	// Handle dash input
	if r.IsKeyPressed(r.KeySpace) && p.DashCooldownTimer <= 0 && !p.IsDashing {
		p.IsDashing = true
		p.DashTimer = p.DashDuration
		p.DashCooldownTimer = p.DashCooldown
	}

	// Handle dash movement
	if p.IsDashing {
		p.DashTimer -= deltaTime
		p.X += p.LastMoveDirection.X * p.DashSpeed * deltaTime * 60
		p.Y += p.LastMoveDirection.Y * p.DashSpeed * deltaTime * 60

		if p.DashTimer <= 0 {
			p.IsDashing = false
		}

		// Add ghost position
		if len(p.GhostTrail) >= p.GhostTrailLength {
			p.GhostTrail = p.GhostTrail[1:]
		}
		p.GhostTrail = append(p.GhostTrail, struct {
			Position   r.Vector2
			Alpha      float32
			FacingLeft bool
		}{
			Position:   r.Vector2{X: p.X, Y: p.Y},
			Alpha:      0.7,
			FacingLeft: p.FacingLeft,
		})
	} else {
		// Fade out ghost trail when not dashing
		if len(p.GhostTrail) > 0 {
			for i := range p.GhostTrail {
				p.GhostTrail[i].Alpha -= 0.1
			}
			// Remove fully faded ghosts
			newTrail := make([]struct {
				Position   r.Vector2
				Alpha      float32
				FacingLeft bool
			}, 0)
			for _, ghost := range p.GhostTrail {
				if ghost.Alpha > 0 {
					newTrail = append(newTrail, ghost)
				}
			}
			p.GhostTrail = newTrail
		}

		// Normal movement code
		nextX := p.X
		nextY := p.Y
		isMoving := false

		// Track movement direction for dash
		if r.IsKeyDown(r.KeyA) {
			nextX -= p.Speed * deltaTime * 60
			isMoving = true
			p.FacingLeft = true
			p.LastMoveDirection = r.Vector2{X: -1, Y: 0}
		}
		if r.IsKeyDown(r.KeyD) {
			nextX += p.Speed * deltaTime * 60
			isMoving = true
			p.FacingLeft = false
			p.LastMoveDirection = r.Vector2{X: 1, Y: 0}
		}
		if r.IsKeyDown(r.KeyW) {
			nextY -= p.Speed * deltaTime * 60
			isMoving = true
			p.LastMoveDirection = r.Vector2{X: 0, Y: -1}
		}
		if r.IsKeyDown(r.KeyS) {
			nextY += p.Speed * deltaTime * 60
			isMoving = true
			p.LastMoveDirection = r.Vector2{X: 0, Y: 1}
		}

		// Keep player within bounds
		p.X = float32(Max(0, Min(float64(nextX), float64(p.GameWidth-p.Width))))
		p.Y = float32(Max(0, Min(float64(nextY), float64(p.GameHeight-p.Height))))

		// Update animation
		p.IsMoving = isMoving
		if p.IsMoving {
			p.FrameTime += deltaTime
			if p.FrameTime >= 0.2 {
				p.FrameTime = 0
				p.CurrentFrame = (p.CurrentFrame + 1) % p.FrameCount
			}
		} else {
			p.CurrentFrame = 3
			p.FrameTime = 0
		}
	}

	// Update invincibility timer
	if p.InvincibleTimer > 0 {
		p.InvincibleTimer -= deltaTime
	}

	// Handle health regeneration
	if p.CurrentHealth < p.MaxHealth {
		p.RegenTimer += deltaTime
		if p.RegenTimer >= p.RegenInterval {
			p.RegenTimer = 0
			p.Heal(1) // Heal 1 HP every 2 seconds
		}
	}
}

// Draw renders the player
func (p *Player) Draw(debug bool, camera r.Camera2D) {
	// Draw ghost trail
	for _, ghost := range p.GhostTrail {
		frameRect := r.Rectangle{
			X:      float32(p.CurrentFrame) * float32(p.FrameWidth),
			Y:      0,
			Width:  float32(p.FrameWidth),
			Height: float32(p.FrameHeight),
		}
		if ghost.FacingLeft {
			frameRect.Width = -frameRect.Width
		}

		r.DrawTexturePro(
			p.Texture,
			frameRect,
			r.Rectangle{
				X:      ghost.Position.X,
				Y:      ghost.Position.Y,
				Width:  float32(p.Width),
				Height: float32(p.Height),
			},
			r.Vector2{X: 0, Y: 0},
			0,
			r.ColorAlpha(r.White, ghost.Alpha),
		)
	}

	// Calculate source rectangle for current frame
	sourceRec := r.Rectangle{
		X:      float32(p.CurrentFrame * p.FrameWidth),
		Y:      0,
		Width:  float32(p.FrameWidth),
		Height: float32(p.FrameHeight),
	}
	if p.FacingLeft {
		sourceRec.Width = -sourceRec.Width
	}

	// Calculate destination rectangle
	destRec := r.Rectangle{
		X:      p.X,
		Y:      p.Y,
		Width:  float32(p.Width),
		Height: float32(p.Height),
	}

	// Set origin to center of sprite
	origin := r.Vector2{
		X: float32(p.Width) / 2,
		Y: float32(p.Height) / 2,
	}

	// Adjust destination X to compensate for centered origin
	destRec.X += float32(p.Width) / 2
	destRec.Y += float32(p.Height) / 2

	// Calculate alpha for blinking effect during invincibility
	alpha := uint8(255)
	if p.InvincibleTimer > 0 {
		// Blink rapidly during invincibility
		if int(p.InvincibleTimer*10)%2 == 0 {
			alpha = 128
		}
	}

	// Draw the current frame with alpha
	r.DrawTexturePro(
		p.Texture,
		sourceRec,
		destRec,
		origin,
		0,
		r.Color{R: 255, G: 255, B: 255, A: alpha},
	)

	// Debug: Draw collision box only when debug is true
	if debug {
		r.DrawRectangleLines(
			int32(p.X),
			int32(p.Y),
			p.Width,
			p.Height,
			r.Red,
		)
	}

	// Draw slash animation if active
	if p.SlashAnim.Active {
		p.SlashAnim.FrameTime += r.GetFrameTime()
		if p.SlashAnim.FrameTime >= 0.05 {
			p.SlashAnim.Frame++
			p.SlashAnim.FrameTime = 0
			if p.SlashAnim.Frame >= 6 {
				p.SlashAnim.Active = false
				p.SlashAnim.Frame = 0
			}
		}

		// Draw the current frame with 0.5 scale
		r.DrawTexturePro(
			p.SlashAnim.Texture,
			r.Rectangle{
				X:      float32(p.SlashAnim.Frame) * float32(p.SlashAnim.FrameWidth),
				Y:      0,
				Width:  float32(p.SlashAnim.FrameWidth),
				Height: float32(p.SlashAnim.FrameHeight),
			},
			r.Rectangle{
				X:      p.X + float32(p.Width)/2 - (float32(p.SlashAnim.FrameWidth)*0.5)/2,
				Y:      p.Y + float32(p.Height)/2 - (float32(p.SlashAnim.FrameHeight)*0.5)/2,
				Width:  float32(p.SlashAnim.FrameWidth) * 0.5,
				Height: float32(p.SlashAnim.FrameHeight) * 0.5,
			},
			r.Vector2{X: 0, Y: 0},
			0,
			r.White,
		)
	}

	// Draw current weapon
	p.CurrentWeapon.Draw(p, camera, debug)
}

// Unload frees the texture from memory
func (p *Player) Unload() {
	r.UnloadTexture(p.Texture)
	r.UnloadTexture(p.SlashAnim.Texture)
}

// GetBounds returns the player's bounding rectangle
func (p *Player) GetBounds() r.Rectangle {
	return r.Rectangle{
		X:      p.X,
		Y:      p.Y,
		Width:  float32(p.Width),
		Height: float32(p.Height),
	}
}

// Helper functions for min/max operations
func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Update CheckRayCollision method
func (p *Player) CheckRayCollision(enemy *Enemy) bool {
	if raygun, ok := p.CurrentWeapon.(*RayGun); ok {
		return raygun.CheckRayCollision(p, enemy.GetBounds()) // Use enemy.GetBounds() instead of enemy
	}
	return false
}

// Add TakeDamage method to Player struct
func (p *Player) TakeDamage(damage int32) {
	if p.InvincibleTimer <= 0 { // Only take damage if not invincible
		p.CurrentHealth -= damage
		if p.CurrentHealth <= 0 {
			p.CurrentHealth = 0
		}
		p.InvincibleTimer = p.InvincibleTime // Start invincibility period
	}
}

// Add IsDead method to Player struct
func (p *Player) IsDead() bool {
	return p.CurrentHealth <= 0
}

// Add Heal method to Player
func (p *Player) Heal(amount int32) {
	newHealth := p.CurrentHealth + amount
	if newHealth > p.MaxHealth {
		newHealth = p.MaxHealth
	}
	p.CurrentHealth = newHealth
}

// Add method to switch weapons
func (p *Player) SwitchWeapon(index int) {
	if index >= 0 && index < len(p.Weapons) {
		p.CurrentWeapon = p.Weapons[index]
	}
}

// Add method to gain experience
func (p *Player) GainExperience(amount int) {
	p.Experience += amount
	if p.Experience >= p.NextLevelExp {
		p.LevelUp()
	}
}

// Add method for leveling up
func (p *Player) LevelUp() {
	p.Experience -= p.NextLevelExp
	p.NextLevelExp = int(float32(p.NextLevelExp) * 1.5) // Increase required exp by 50%
	p.MaxHealth += 10
	p.CurrentHealth = p.MaxHealth
	// You can add more level-up bonuses here
}

// Add this method to Player struct
func (p *Player) UpdateHarvestDamage(inventory *Inventory) {
	// Base harvest damage
	p.HarvestDamage = 1

	// Check if inventory exists and has pickaxe
	if inventory != nil {
		if count, exists := inventory.ItemCounts["Pickaxe"]; exists && count > 0 {
			p.HarvestDamage += 3 // Add 3 harvest damage if pickaxe is in inventory
		}
	}
}
