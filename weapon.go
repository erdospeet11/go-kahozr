package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Weapon is the base interface for all weapons
type Weapon interface {
	Update(deltaTime float32, player *Player)
	Draw(player *Player, camera rl.Camera2D, debug bool)
	OnActivate(player *Player, camera rl.Camera2D)
	OnDeactivate(player *Player)
	IsActive() bool
}

// BaseWeapon contains common weapon properties
type BaseWeapon struct {
	IsEquipped bool
	Active     bool
}

// RayGun implements the Weapon interface
type RayGun struct {
	BaseWeapon
	HeatLevel    float32
	IsOverheated bool
	CooldownRate float32
	HeatRate     float32
	rayDirection rl.Vector2
	AimLength    float32
	Texture      rl.Texture2D
}

func NewRayGun() *RayGun {
	return &RayGun{
		BaseWeapon: BaseWeapon{
			IsEquipped: true,
			Active:     false,
		},
		HeatLevel:    0,
		IsOverheated: false,
		CooldownRate: 30.0,
		HeatRate:     40.0,
		AimLength:    100.0,
		Texture:      rl.LoadTexture("assets/ray-gun.png"),
	}
}

func (r *RayGun) Update(deltaTime float32, player *Player) {
	// Handle heat mechanics
	if r.Active {
		r.HeatLevel += r.HeatRate * deltaTime
		if r.HeatLevel >= 100 {
			r.HeatLevel = 100
			r.IsOverheated = true
			r.Active = false
		}
	} else {
		r.HeatLevel -= r.CooldownRate * deltaTime
		if r.HeatLevel <= 0 {
			r.HeatLevel = 0
			r.IsOverheated = false
		}
	}
}

func (r *RayGun) Draw(player *Player, camera rl.Camera2D, debug bool) {
	if !r.Active {
		return
	}

	mouseScreen := rl.GetMousePosition()
	mouseWorld := rl.GetScreenToWorld2D(mouseScreen, camera)

	playerCenter := rl.Vector2{
		X: player.X + float32(player.Width)/2,
		Y: player.Y + float32(player.Height)/2,
	}

	direction := rl.Vector2{
		X: mouseWorld.X - playerCenter.X,
		Y: mouseWorld.Y - playerCenter.Y,
	}

	length := float32(Sqrt(float64(direction.X*direction.X + direction.Y*direction.Y)))
	if length > 0 {
		normalizedDir := rl.Vector2{
			X: direction.X / length,
			Y: direction.Y / length,
		}

		rayLength := float32(Min(float64(length), float64(r.AimLength)))
		endPoint := rl.Vector2{
			X: playerCenter.X + normalizedDir.X*rayLength,
			Y: playerCenter.Y + normalizedDir.Y*rayLength,
		}

		// Store ray direction for collision checks
		r.rayDirection = rl.Vector2{
			X: normalizedDir.X * rayLength,
			Y: normalizedDir.Y * rayLength,
		}

		rl.DrawLineEx(playerCenter, endPoint, 2, rl.Red)
	}

	// Draw heat bar when active
	if r.Active || r.HeatLevel > 0 {
		barWidth := float32(30)
		barHeight := float32(4)
		barX := player.X + float32(player.Width)/2 - barWidth/2
		barY := player.Y - barHeight - 2

		// Background
		rl.DrawRectangleV(
			rl.Vector2{X: barX, Y: barY},
			rl.Vector2{X: barWidth, Y: barHeight},
			rl.DarkGray,
		)

		// Heat level
		heatColor := rl.Yellow
		if r.IsOverheated {
			heatColor = rl.Red
		}
		rl.DrawRectangleV(
			rl.Vector2{X: barX, Y: barY},
			rl.Vector2{X: barWidth * (r.HeatLevel / 100), Y: barHeight},
			heatColor,
		)
	}
}

func (r *RayGun) OnActivate(player *Player, camera rl.Camera2D) {
	if !r.IsOverheated {
		r.Active = true
	}
}

func (r *RayGun) OnDeactivate(player *Player) {
	r.Active = false
}

func (r *RayGun) IsActive() bool {
	return r.Active
}

func (r *RayGun) CheckRayCollision(player *Player, bounds rl.Rectangle) bool {
	if !r.Active {
		return false
	}

	playerCenter := rl.Vector2{
		X: player.X + float32(player.Width)/2,
		Y: player.Y + float32(player.Height)/2,
	}

	// Check multiple points along the ray
	for t := float32(0); t <= 1.0; t += 0.1 {
		point := rl.Vector2{
			X: playerCenter.X + r.rayDirection.X*t,
			Y: playerCenter.Y + r.rayDirection.Y*t,
		}
		if rl.CheckCollisionPointRec(point, bounds) {
			return true
		}
	}
	return false
}

func (r *RayGun) Unload() {
	rl.UnloadTexture(r.Texture)
}

// Add Sword struct after RayGun
type Sword struct {
	BaseWeapon
	Texture    rl.Texture2D
	IsSlashing bool
	SlashAnim  struct {
		Texture     rl.Texture2D
		Frame       int32
		FrameTime   float32
		FrameWidth  int32
		FrameHeight int32
	}
	DamageArea rl.Rectangle
}

func NewSword() *Sword {
	sword := &Sword{
		BaseWeapon: BaseWeapon{
			IsEquipped: false,
			Active:     false,
		},
		Texture: rl.LoadTexture("assets/sword.png"),
	}

	sword.SlashAnim = struct {
		Texture     rl.Texture2D
		Frame       int32
		FrameTime   float32
		FrameWidth  int32
		FrameHeight int32
	}{
		Texture:     rl.LoadTexture("assets/slash.png"),
		FrameWidth:  65,
		FrameHeight: 27,
	}

	return sword
}

func (s *Sword) Update(deltaTime float32, player *Player) {
	if s.IsSlashing {
		s.SlashAnim.FrameTime += deltaTime
		if s.SlashAnim.FrameTime >= 0.05 {
			s.SlashAnim.Frame++
			s.SlashAnim.FrameTime = 0
			if s.SlashAnim.Frame >= 6 {
				s.IsSlashing = false
				s.SlashAnim.Frame = 0
				s.Active = false
			}
		}
	}
}

func (s *Sword) Draw(player *Player, camera rl.Camera2D, debug bool) {
	if s.IsSlashing {
		// Update damage area position based on facing direction
		if player.FacingLeft {
			s.DamageArea = rl.Rectangle{
				X:      player.X - float32(s.SlashAnim.FrameWidth)*0.5,
				Y:      player.Y - float32(player.Height)/2,
				Width:  float32(s.SlashAnim.FrameWidth) * 0.5,
				Height: float32(s.SlashAnim.FrameHeight),
			}
		} else {
			s.DamageArea = rl.Rectangle{
				X:      player.X + float32(player.Width),
				Y:      player.Y - float32(player.Height)/2,
				Width:  float32(s.SlashAnim.FrameWidth) * 0.5,
				Height: float32(s.SlashAnim.FrameHeight),
			}
		}

		centerX := player.X + float32(player.Width)/2
		centerY := player.Y + float32(player.Height)/2

		// Draw slash animation with flipping
		rl.DrawTexturePro(
			s.SlashAnim.Texture,
			rl.Rectangle{
				X:      float32(s.SlashAnim.Frame)*float32(s.SlashAnim.FrameWidth) + 1,
				Y:      0,
				Width:  (float32(s.SlashAnim.FrameWidth) - 2) * float32(map[bool]int{true: -1, false: 1}[player.FacingLeft]),
				Height: float32(s.SlashAnim.FrameHeight),
			},
			rl.Rectangle{
				X:      centerX,
				Y:      centerY,
				Width:  float32(s.SlashAnim.FrameWidth),
				Height: float32(s.SlashAnim.FrameHeight),
			},
			rl.Vector2{X: float32(s.SlashAnim.FrameWidth) / 2, Y: float32(s.SlashAnim.FrameHeight) / 2},
			0,
			rl.White,
		)

		// Draw damage area in debug mode
		if debug {
			rl.DrawRectangleLinesEx(s.DamageArea, 1, rl.Red)
		}
	}
}

func (s *Sword) OnActivate(player *Player, camera rl.Camera2D) {
	if !s.IsSlashing {
		s.IsSlashing = true
		s.Active = true
		s.SlashAnim.Frame = 0
		s.SlashAnim.FrameTime = 0
	}
}

func (s *Sword) OnDeactivate(player *Player) {
	// Nothing needed here
}

func (s *Sword) IsActive() bool {
	return s.Active
}

func (s *Sword) Unload() {
	rl.UnloadTexture(s.Texture)
	rl.UnloadTexture(s.SlashAnim.Texture)
}

// Add method to check for sword collision
func (s *Sword) CheckSlashCollision(bounds rl.Rectangle) bool {
	if !s.IsSlashing {
		return false
	}
	return rl.CheckCollisionRecs(s.DamageArea, bounds)
}

// Update Dummy struct
type Dummy struct {
	X              float32
	Y              float32
	Width          int32
	Height         int32
	Texture        rl.Texture2D
	DamageCooldown float32
	DamageText     struct {
		Value    int32
		Position rl.Vector2
		Alpha    float32
		Timer    float32
	}
}

// Update NewDummy to initialize cooldown
func NewDummy(x, y float32) *Dummy {
	return &Dummy{
		X:              x,
		Y:              y,
		Width:          16,
		Height:         32,
		Texture:        rl.LoadTexture("assets/dummy.png"),
		DamageCooldown: 0,
	}
}

// Add Update method for Dummy
func (d *Dummy) Update() {
	if d.DamageCooldown > 0 {
		d.DamageCooldown -= rl.GetFrameTime()
	}
}

func (d *Dummy) Draw(debug bool) {
	// Draw dummy sprite
	rl.DrawTextureEx(
		d.Texture,
		rl.Vector2{X: d.X, Y: d.Y},
		0,
		1,
		rl.White,
	)

	// Draw damage text if active
	if d.DamageText.Timer > 0 {
		text := fmt.Sprintf("-%d", d.DamageText.Value)
		rl.DrawText(
			text,
			int32(d.DamageText.Position.X),
			int32(d.DamageText.Position.Y),
			14,
			rl.ColorAlpha(rl.Red, d.DamageText.Alpha),
		)

		// Update position and alpha
		d.DamageText.Position.Y -= 1
		d.DamageText.Timer -= rl.GetFrameTime()
		d.DamageText.Alpha = d.DamageText.Timer
	}
}

func (d *Dummy) GetBounds() rl.Rectangle {
	return rl.Rectangle{
		X:      d.X,
		Y:      d.Y,
		Width:  float32(d.Width),
		Height: float32(d.Height),
	}
}

func (d *Dummy) TakeDamage(damage int32) {
	// Update damage text
	d.DamageText = struct {
		Value    int32
		Position rl.Vector2
		Alpha    float32
		Timer    float32
	}{
		Value: damage,
		Position: rl.Vector2{
			X: d.X + float32(d.Width)/2 - 8,
			Y: d.Y - 10,
		},
		Alpha: 1.0,
		Timer: 1.0,
	}
}

func (d *Dummy) Unload() {
	rl.UnloadTexture(d.Texture)
}

// Add Pistol struct
type Pistol struct {
	BaseWeapon
	Texture rl.Texture2D
	Bullets []struct {
		Position  rl.Vector2
		Direction rl.Vector2
		Speed     float32
		LifeTimer float32
	}
	ShootCooldown float32
	CooldownTimer float32
}

func NewPistol() *Pistol {
	return &Pistol{
		BaseWeapon: BaseWeapon{
			IsEquipped: true,
			Active:     false,
		},
		Texture: rl.LoadTexture("assets/pistol.png"),
		Bullets: make([]struct {
			Position  rl.Vector2
			Direction rl.Vector2
			Speed     float32
			LifeTimer float32
		}, 0),
		ShootCooldown: 0.15, // Slightly faster firing rate
	}
}

func (p *Pistol) Update(deltaTime float32, player *Player) {
	// Update cooldown
	if p.CooldownTimer > 0 {
		p.CooldownTimer -= deltaTime
	}

	// Update bullets
	var remainingBullets []struct {
		Position  rl.Vector2
		Direction rl.Vector2
		Speed     float32
		LifeTimer float32
	}

	for _, bullet := range p.Bullets {
		// Update bullet position
		newBullet := bullet
		newBullet.Position.X += bullet.Direction.X * bullet.Speed * deltaTime
		newBullet.Position.Y += bullet.Direction.Y * bullet.Speed * deltaTime
		newBullet.LifeTimer -= deltaTime

		if newBullet.LifeTimer > 0 {
			remainingBullets = append(remainingBullets, newBullet)
		}
	}
	p.Bullets = remainingBullets
}

func (p *Pistol) Draw(player *Player, camera rl.Camera2D, debug bool) {
	// Draw bullets
	for _, bullet := range p.Bullets {
		rl.DrawRectangle(
			int32(bullet.Position.X-2),
			int32(bullet.Position.Y-2),
			4, 4,
			rl.Yellow,
		)
	}
}

func (p *Pistol) OnActivate(player *Player, camera rl.Camera2D) {
	if p.CooldownTimer <= 0 {
		// Get mouse position in world space
		mouseScreen := rl.GetMousePosition()
		mouseWorld := rl.GetScreenToWorld2D(mouseScreen, camera)

		// Calculate direction
		playerCenter := rl.Vector2{
			X: player.X + float32(player.Width)/2,
			Y: player.Y + float32(player.Height)/2,
		}
		direction := rl.Vector2{
			X: mouseWorld.X - playerCenter.X,
			Y: mouseWorld.Y - playerCenter.Y,
		}

		// Normalize direction
		length := float32(math.Sqrt(float64(direction.X*direction.X + direction.Y*direction.Y)))
		if length > 0 {
			direction.X /= length
			direction.Y /= length
		}

		// Create new bullet
		p.Bullets = append(p.Bullets, struct {
			Position  rl.Vector2
			Direction rl.Vector2
			Speed     float32
			LifeTimer float32
		}{
			Position:  playerCenter,
			Direction: direction,
			Speed:     240.0,
			LifeTimer: 0.5,
		})

		p.CooldownTimer = p.ShootCooldown
	}
}

func (p *Pistol) OnDeactivate(player *Player) {
	// Nothing needed here
}

func (p *Pistol) IsActive() bool {
	return p.Active
}

func (p *Pistol) CheckBulletCollision(bounds rl.Rectangle) bool {
	for _, bullet := range p.Bullets {
		bulletBounds := rl.Rectangle{
			X:      bullet.Position.X - 2,
			Y:      bullet.Position.Y - 2,
			Width:  4,
			Height: 4,
		}
		if rl.CheckCollisionRecs(bulletBounds, bounds) {
			return true
		}
	}
	return false
}

func (p *Pistol) Unload() {
	rl.UnloadTexture(p.Texture)
}
