package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	r "github.com/gen2brain/raylib-go/raylib"
)

const (
	GameWidth  = 900
	GameHeight = 800
)

// GameState represents the current state of the game
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StateDead
)

// Camera represents the game camera
type Camera struct {
	Target   r.Vector2
	Offset   r.Vector2
	Rotation float32
	Zoom     float32
}

// Game represents the main game state and objects
type Game struct {
	state            GameState
	menu             *MainMenu
	camera           Camera
	player           *Player
	enemies          []*Enemy
	sprites          []*Sprite
	trees            []*Tree
	droppedItems     []*DroppedItem
	inventory        *Inventory
	debug            bool
	cursorTex        r.Texture2D
	stones           []*Stone
	gameFont         r.Font
	toolbarSlots     []r.Rectangle
	gameTimer        float32
	crafting         *CraftingSystem
	particles        *ParticleSystem
	portals          []*Portal
	portalSpawnTimer float32
	shakeAmount      float32
	shakeTimer       float32
	merchant         *Merchant
	isPaused         bool
	dummies          []*Dummy
}

// NewGame creates a new game instance
func NewGame() *Game {
	game := &Game{
		state: StateMenu,
		camera: Camera{
			Target:   r.Vector2{X: 0, Y: 0},
			Offset:   r.Vector2{X: 400, Y: 300},
			Rotation: 0,
			Zoom:     3.0,
		},
		debug:            false,
		crafting:         NewCraftingSystem(),
		particles:        NewParticleSystem(),
		portals:          make([]*Portal, 0),
		portalSpawnTimer: 10.0,
	}

	// Create menu after font is loaded
	game.menu = NewMainMenu(game.gameFont)

	return game
}

// Initialize sets up the game
func (g *Game) Initialize() {
	r.InitWindow(800, 600, "POKET KOZMOZ")
	r.SetTargetFPS(60)

	// Set window icon
	icon := r.LoadImage("assets/pulze.png")
	r.SetWindowIcon(*icon)
	r.UnloadImage(icon)

	// Hide default cursor
	r.HideCursor()

	// Load cursor texture
	g.cursorTex = r.LoadTexture("assets/cursor.png")

	// Use default font
	g.gameFont = r.GetFontDefault()

	// Initialize toolbar slots
	slotSize := int32(50)
	spacing := int32(10)
	startX := (800 - (5*slotSize + 4*spacing)) / 2
	startY := 600 - slotSize - 10

	g.toolbarSlots = make([]r.Rectangle, 5)
	for i := range g.toolbarSlots {
		g.toolbarSlots[i] = r.Rectangle{
			X:      float32(startX + int32(i)*(slotSize+spacing)),
			Y:      float32(startY),
			Width:  float32(slotSize),
			Height: float32(slotSize),
		}
	}

	// Initialize menu after window is created
	g.menu = NewMainMenu(g.gameFont)
	g.menu.Initialize()
}

// InitializeGameObjects creates game objects when starting to play
func (g *Game) InitializeGameObjects() {
	g.player = NewPlayer(400, 300, GameWidth, GameHeight)
	g.inventory = NewInventory(g.player)
	g.enemies = make([]*Enemy, 0)

	// Create sprites
	g.sprites = make([]*Sprite, 20)
	for i := range g.sprites {
		g.sprites[i] = NewSprite("assets/grass.png")
	}

	// Create trees with collision check
	g.trees = make([]*Tree, 15)
	for i := range g.trees {
		var tree *Tree
		for {
			tree = NewTree(GameWidth, GameHeight)
			bounds := r.Rectangle{
				X:      tree.X,
				Y:      tree.Y,
				Width:  float32(tree.Width),
				Height: float32(tree.Height),
			}
			if !g.IsPositionOccupied(bounds, 20) {
				break
			}
		}
		g.trees[i] = tree
	}

	// Create stones with collision check
	g.stones = make([]*Stone, 25)
	for i := range g.stones {
		var stone *Stone
		for {
			stone = NewStone(GameWidth, GameHeight)
			bounds := r.Rectangle{
				X:      stone.X,
				Y:      stone.Y,
				Width:  float32(stone.Width),
				Height: float32(stone.Height),
			}
			if !g.IsPositionOccupied(bounds, 20) {
				break
			}
		}
		g.stones[i] = stone
	}

	g.loadItemIcon("Goodie Bag", "assets/goodie-bag.png")

	// Create merchant at a fixed position
	g.merchant = NewMerchant(500, 200)
	g.merchant.LoadIcons()

	// Create a dummy
	g.dummies = append(g.dummies, NewDummy(400, 400))

	// Add icon loading here
	g.inventory.LoadIcon("Pickaxe", "assets/pickaxe.png")
	g.inventory.LoadIcon("Strange Log", "assets/tree-pickup.png")
	g.inventory.LoadIcon("Stone Fragment", "assets/stone-pickup.png")
	g.inventory.LoadIcon("Golden Nugget", "assets/gold-nugget.png")
	g.inventory.LoadIcon("Gold Coin", "assets/gold_coin.png")
}

// Update handles game logic updates
func (g *Game) Update() {
	switch g.state {
	case StateMenu:
		if g.menu.Update() {
			g.InitializeGameObjects()
			g.state = StatePlaying
		}

	case StatePlaying:
		// Check for player death
		if g.player != nil && g.player.IsDead() {
			g.state = StateDead
		}
		g.UpdatePlaying()

	case StateDead:
		// Click anywhere to return to menu
		if r.IsMouseButtonPressed(0) {
			g.state = StateMenu
			g.menu.Initialize()
		}
	}

	// Only update harvest damage if player and inventory exist
	if g.player != nil && g.inventory != nil {
		g.player.UpdateHarvestDamage(g.inventory)
	}
}

// UpdatePlaying handles updates during gameplay
func (g *Game) UpdatePlaying() {
	// Handle window toggles and pausing
	if r.IsKeyPressed(r.KeyE) {
		g.inventory.IsOpen = !g.inventory.IsOpen
		g.crafting.IsOpen = false
		g.merchant.IsOpen = false
		g.isPaused = g.inventory.IsOpen
	}

	if r.IsKeyPressed(r.KeyC) {
		g.crafting.IsOpen = !g.crafting.IsOpen
		g.inventory.IsOpen = false
		g.merchant.IsOpen = false
		g.isPaused = g.crafting.IsOpen
	}

	// Update merchant interaction
	if g.merchant != nil && g.merchant.IsOpen {
		g.inventory.IsOpen = false
		g.crafting.IsOpen = false
		g.isPaused = true
	}

	// Handle merchant clicking
	if r.IsMouseButtonPressed(0) {
		screenPos := r.GetMousePosition()
		worldPos := r.GetScreenToWorld2D(screenPos, r.Camera2D{
			Target:   g.camera.Target,
			Offset:   g.camera.Offset,
			Rotation: g.camera.Rotation,
			Zoom:     g.camera.Zoom,
		})

		if !g.merchant.IsOpen && g.merchant.OnClick(worldPos) {
			g.merchant.IsOpen = true
			g.inventory.IsOpen = false
			g.crafting.IsOpen = false
			g.isPaused = true
		}
	}

	// Fix merchant close handling
	if g.merchant != nil {
		if !g.merchant.IsOpen { // Merchant was just closed
			g.isPaused = false
		}
	}

	// Only update game logic if not paused
	if !g.isPaused {
		g.UpdateGameLogic()
	}

	// Always update UI-related things
	g.UpdateUI()

	// Add weapon switching with number keys
	if r.IsKeyPressed(r.KeyOne) {
		g.player.SwitchWeapon(0)
	}
	if r.IsKeyPressed(r.KeyTwo) {
		g.player.SwitchWeapon(1)
	}
	if r.IsKeyPressed(r.KeyThree) {
		g.player.SwitchWeapon(2)
	}
}

// UpdateGameLogic moves all the non-UI game update logic here
func (g *Game) UpdateGameLogic() {
	if g.player != nil {
		g.player.Update(r.Camera2D{
			Target:   g.camera.Target,
			Offset:   g.camera.Offset,
			Rotation: g.camera.Rotation,
			Zoom:     g.camera.Zoom,
		})
	}

	// Update portal spawn timer
	g.portalSpawnTimer -= r.GetFrameTime()
	if g.portalSpawnTimer <= 0 {
		g.portalSpawnTimer = 10.0 // Reset timer for next portal
		// Pass game instance to NewPortal
		g.portals = append(g.portals, NewPortal(GameWidth, GameHeight, g))
	}

	// Update all portals and spawn enemies
	var remainingPortals []*Portal
	for _, portal := range g.portals {
		if portal.Update(r.GetFrameTime()) {
			spawnPos := portal.GetSpawnPosition()
			newEnemy := NewEnemy(spawnPos.X, spawnPos.Y, g.player)
			g.enemies = append(g.enemies, newEnemy)
		}

		// Keep portal if not done
		if !portal.IsDone {
			remainingPortals = append(remainingPortals, portal)
		} else {
			portal.Unload()
		}
	}
	g.portals = remainingPortals

	// Update all enemies
	var remainingEnemies []*Enemy
	for _, enemy := range g.enemies {
		enemy.Update()

		// Check for enemy-player collision and damage
		if g.player != nil && enemy.CheckCollision(g.player) {
			g.player.TakeDamage(10)
			g.particles.SpawnExplosion(r.Red, 10, g.player.X, g.player.Y)
			g.shakeAmount = 3.0
			g.shakeTimer = 0.1
		}

		// Check for weapon damage
		if g.player != nil {
			if raygun, ok := g.player.CurrentWeapon.(*RayGun); ok && raygun.Active && !raygun.IsOverheated {
				if raygun.CheckRayCollision(g.player, enemy.GetBounds()) {
					if enemy.DamageCooldown <= 0 {
						damagePerSecond := int32(3)
						enemy.TakeDamage(1)
						enemy.DamageCooldown = 1.0 / float32(damagePerSecond)
						g.particles.SpawnExplosion(r.Yellow, 10, enemy.X, enemy.Y)
						g.shakeAmount = 3.0
						g.shakeTimer = 0.1
					}
				}
			} else if sword, ok := g.player.CurrentWeapon.(*Sword); ok {
				if sword.CheckSlashCollision(enemy.GetBounds()) {
					if enemy.DamageCooldown <= 0 {
						enemy.TakeDamage(2)        // Reduced from 5 to 2 damage
						enemy.DamageCooldown = 0.3 // Slightly faster than before
						g.particles.SpawnExplosion(r.White, 10, enemy.X, enemy.Y)
						g.shakeAmount = 3.0
						g.shakeTimer = 0.1
					}
				}
			} else if pistol, ok := g.player.CurrentWeapon.(*Pistol); ok {
				if pistol.CheckBulletCollision(enemy.GetBounds()) {
					if enemy.DamageCooldown <= 0 {
						enemy.TakeDamage(2)
						enemy.DamageCooldown = 0.2
						g.particles.SpawnExplosion(r.Yellow, 5, enemy.X, enemy.Y)
						g.shakeAmount = 2.0
						g.shakeTimer = 0.05
					}
				}
			}
		}

		// Check if enemy is dead
		if enemy.IsDead() {
			g.particles.SpawnExplosion(r.Red, 15, enemy.X, enemy.Y)
			if rand.Float32() < enemy.DropChance {
				dropPos := enemy.GetDropPosition()
				g.droppedItems = append(g.droppedItems, NewDroppedItem(
					dropPos.X-8,
					dropPos.Y-8,
					"assets/goodie-bag.png",
					"Goodie Bag",
				))
			}
			enemy.Unload()

			// Give player experience
			g.player.GainExperience(10) // Adjust experience amount as needed
		} else {
			remainingEnemies = append(remainingEnemies, enemy)
		}
	}
	g.enemies = remainingEnemies

	// Toggle debug with F1
	if r.IsKeyPressed(r.KeyF1) {
		g.debug = !g.debug
	}

	// Handle zoom
	wheel := r.GetMouseWheelMove()
	if wheel != 0 {
		g.camera.Zoom = float32(math.Max(1.0, math.Min(3.0, float64(g.camera.Zoom+wheel*0.1))))
	}

	// Update camera position
	g.UpdateCamera()

	// Handle tree clicking
	if r.IsMouseButtonPressed(0) {
		screenPos := r.GetMousePosition()
		worldPos := r.GetScreenToWorld2D(screenPos, r.Camera2D{
			Target:   g.camera.Target,
			Offset:   g.camera.Offset,
			Rotation: g.camera.Rotation,
			Zoom:     g.camera.Zoom,
		})

		g.HandleTreeClicks(worldPos)
	}

	// Handle stone clicking
	if r.IsMouseButtonPressed(0) {
		screenPos := r.GetMousePosition()
		worldPos := r.GetScreenToWorld2D(screenPos, r.Camera2D{
			Target:   g.camera.Target,
			Offset:   g.camera.Offset,
			Rotation: g.camera.Rotation,
			Zoom:     g.camera.Zoom,
		})

		g.HandleStoneClicks(worldPos)
	}

	// Check for item pickups
	g.CheckItemPickups()

	// Update timer
	g.gameTimer += r.GetFrameTime()

	// Update particles
	g.particles.Update()

	// Spawn explosion on F2
	if r.IsKeyPressed(r.KeyF2) && g.player != nil {
		centerX := g.player.X + float32(g.player.Width)/2
		centerY := g.player.Y + float32(g.player.Height)/2
		g.particles.SpawnExplosion(r.Red, 25, centerX, centerY)
	}

	// Update the shake timer
	if g.shakeTimer > 0 {
		g.shakeTimer -= r.GetFrameTime()
		if g.shakeTimer <= 0 {
			g.shakeAmount = 0
		}
	}

	// Check for ray collision with dummies
	if g.player != nil {
		for _, dummy := range g.dummies {
			dummy.Update()
			if raygun, ok := g.player.CurrentWeapon.(*RayGun); ok {
				if raygun.CheckRayCollision(g.player, dummy.GetBounds()) {
					if raygun.Active && !raygun.IsOverheated && dummy.DamageCooldown <= 0 {
						damagePerSecond := int32(3)
						dummy.TakeDamage(1)
						dummy.DamageCooldown = 1.0 / float32(damagePerSecond)
					}
				}
			}
		}
	}
}

// UpdateCamera updates the camera position
func (g *Game) UpdateCamera() {
	if g.player == nil {
		return
	}

	viewWidth := 800 / g.camera.Zoom
	viewHeight := 600 / g.camera.Zoom
	minX := viewWidth / 2
	minY := viewHeight / 2
	maxX := float32(GameWidth) - (viewWidth / 2)
	maxY := float32(GameHeight) - (viewHeight / 2)

	// Calculate base camera position
	targetX := float32(math.Max(float64(minX),
		math.Min(float64(maxX), float64(g.player.X))))
	targetY := float32(math.Max(float64(minY),
		math.Min(float64(maxY), float64(g.player.Y))))

	// Add random shake offset if shake is active
	if g.shakeTimer > 0 {
		shakeX := (rand.Float32()*2 - 1) * g.shakeAmount
		shakeY := (rand.Float32()*2 - 1) * g.shakeAmount
		targetX += shakeX
		targetY += shakeY
	}

	g.camera.Target.X = targetX
	g.camera.Target.Y = targetY
}

// Update IsPlayerInRange to be more precise
func (g *Game) IsPlayerInRange(targetX, targetY float32, targetWidth, targetHeight int32, range_ float32) bool {
	if g.player == nil {
		return false
	}

	// Calculate center points
	playerCenterX := g.player.X + float32(g.player.Width)/2
	playerCenterY := g.player.Y + float32(g.player.Height)/2
	targetCenterX := targetX + float32(targetWidth)/2
	targetCenterY := targetY + float32(targetHeight)/2

	// Calculate distance
	dx := playerCenterX - targetCenterX
	dy := playerCenterY - targetCenterY
	distSquared := dx*dx + dy*dy

	// Only return true if player is in range AND mouse is over the object
	return distSquared <= range_*range_ && r.CheckCollisionPointRec(
		r.GetScreenToWorld2D(r.GetMousePosition(), r.Camera2D{
			Target:   g.camera.Target,
			Offset:   g.camera.Offset,
			Rotation: g.camera.Rotation,
			Zoom:     g.camera.Zoom,
		}),
		r.Rectangle{
			X:      targetX,
			Y:      targetY,
			Width:  float32(targetWidth),
			Height: float32(targetHeight),
		},
	)
}

// Update HandleTreeClicks method
func (g *Game) HandleTreeClicks(worldPos r.Vector2) {
	var remainingTrees []*Tree
	interactionRange := float32(50)

	for _, tree := range g.trees {
		if g.IsPlayerInRange(tree.X, tree.Y, tree.Width, tree.Height, interactionRange) {
			if tree.OnClick(worldPos, g.player.HarvestDamage) {
				explosionX := tree.X + float32(tree.Width)/2
				explosionY := tree.Y + float32(tree.Height)/2
				g.particles.SpawnExplosion(r.Yellow, 20, explosionX, explosionY)
				// Add dropped item
				g.droppedItems = append(g.droppedItems, NewDroppedItem(
					tree.X+float32(tree.Width-16)/2,
					tree.Y+float32(tree.Height-16)/2,
					"assets/tree-pickup.png",
					"Strange Log",
				))
				tree.Unload()
			} else {
				remainingTrees = append(remainingTrees, tree)
				g.particles.SpawnDamageNumber(
					fmt.Sprintf("-%d", g.player.HarvestDamage),
					tree.X+float32(tree.Width)/2,
					tree.Y-10,
				)
			}
		} else {
			remainingTrees = append(remainingTrees, tree)
		}
	}
	g.trees = remainingTrees
}

// Update HandleStoneClicks method
func (g *Game) HandleStoneClicks(worldPos r.Vector2) {
	var remainingStones []*Stone
	interactionRange := float32(50)

	for _, stone := range g.stones {
		if g.IsPlayerInRange(stone.X, stone.Y, stone.Width, stone.Height, interactionRange) {
			if stone.OnClick(worldPos, g.player.HarvestDamage) {
				explosionX := stone.X + float32(stone.Width)/2
				explosionY := stone.Y + float32(stone.Height)/2
				g.particles.SpawnExplosion(r.Yellow, 20, explosionX, explosionY)
				// Add appropriate dropped item based on stone type
				if stone.IsGolden {
					g.droppedItems = append(g.droppedItems, NewDroppedItem(
						stone.X+float32(stone.Width-16)/2,
						stone.Y+float32(stone.Height-16)/2,
						"assets/gold-nugget.png",
						"Golden Nugget",
					))
				} else {
					g.droppedItems = append(g.droppedItems, NewDroppedItem(
						stone.X+float32(stone.Width-16)/2,
						stone.Y+float32(stone.Height-16)/2,
						"assets/stone-pickup.png",
						"Stone Fragment",
					))
				}
				stone.Unload()
			} else {
				remainingStones = append(remainingStones, stone)
				g.particles.SpawnDamageNumber(
					fmt.Sprintf("-%d", g.player.HarvestDamage),
					stone.X+float32(stone.Width)/2,
					stone.Y-10,
				)
			}
		} else {
			remainingStones = append(remainingStones, stone)
		}
	}
	g.stones = remainingStones
}

// CheckItemPickups handles item collection
func (g *Game) CheckItemPickups() {
	if g.player == nil {
		return
	}

	var remainingItems []*DroppedItem
	for _, item := range g.droppedItems {
		if item.CheckCollision(g.player.GetBounds()) {
			g.inventory.Items = append(g.inventory.Items, item.Name)
			g.inventory.ItemCounts[item.Name]++
			g.loadItemIcon(item.Name, item.ImagePath)
			item.Unload()
		} else {
			remainingItems = append(remainingItems, item)
		}
	}
	g.droppedItems = remainingItems

	// When harvesting trees
	for i, tree := range g.trees {
		if tree != nil && r.CheckCollisionRecs(g.player.GetBounds(), tree.GetBounds()) {
			if r.IsKeyPressed(r.KeyE) {
				tree.Health -= int32(g.player.HarvestDamage)
				if tree.Health <= 0 {
					g.particles.SpawnExplosion(r.Red, 15, tree.X, tree.Y)
					tree.Unload()
					g.trees = append(g.trees[:i], g.trees[i+1:]...)
				}
			}
		}
	}

	// When harvesting stones
	for i, stone := range g.stones {
		if stone != nil && r.CheckCollisionRecs(g.player.GetBounds(), stone.GetBounds()) {
			if r.IsKeyPressed(r.KeyE) {
				stone.Health -= int32(g.player.HarvestDamage)
				if stone.Health <= 0 {
					g.particles.SpawnExplosion(r.Red, 15, stone.X, stone.Y)
					stone.Unload()
					g.stones = append(g.stones[:i], g.stones[i+1:]...)
				}
			}
		}
	}
}

// Draw renders the game
func (g *Game) Draw() {
	r.BeginDrawing()
	r.ClearBackground(r.Color{R: 83, G: 114, B: 133, A: 255})

	switch g.state {
	case StateMenu:
		r.ShowCursor()
		g.menu.Draw()

	case StatePlaying:
		r.HideCursor()

		camera := r.Camera2D{
			Target:   g.camera.Target,
			Offset:   g.camera.Offset,
			Rotation: g.camera.Rotation,
			Zoom:     g.camera.Zoom,
		}

		r.BeginMode2D(camera)

		// Draw game world border
		r.DrawRectangleLines(0, 0, GameWidth, GameHeight, r.DarkGray)

		// Draw all game objects
		for _, portal := range g.portals {
			portal.Draw(g.debug)
		}
		for _, sprite := range g.sprites {
			sprite.Draw()
		}
		for _, tree := range g.trees {
			tree.Draw(g.debug)
		}
		for _, stone := range g.stones {
			stone.Draw(g.debug)
		}
		for _, item := range g.droppedItems {
			item.Draw(g.debug)
		}
		if g.player != nil {
			g.player.Draw(g.debug, camera)
			if sword, ok := g.player.CurrentWeapon.(*Sword); ok {
				sword.Draw(g.player, camera, g.debug)
			}
		}
		for _, enemy := range g.enemies {
			enemy.Draw(g.debug)
		}
		if g.merchant != nil {
			g.merchant.Draw(g.gameFont, g.inventory, camera, g.debug)
		}

		// Draw particles
		g.particles.Draw()

		// Draw dummies
		for _, dummy := range g.dummies {
			dummy.Draw(g.debug)
		}

		r.EndMode2D()

		// Draw UI elements
		g.DrawUI()

	case StateDead:
		// Draw death screen
		r.ClearBackground(r.Black)
		text := "YOU DIED"
		fontSize := int32(60)
		textWidth := r.MeasureText(text, fontSize)
		r.DrawText(text,
			400-textWidth/2,
			250,
			fontSize,
			r.Red,
		)

		clickText := "Click anywhere to continue"
		smallSize := int32(20)
		smallWidth := r.MeasureText(clickText, smallSize)
		r.DrawText(clickText,
			400-smallWidth/2,
			350,
			smallSize,
			r.Red,
		)
	}

	// Draw cursor
	mousePos := r.GetMousePosition()
	r.DrawTextureEx(g.cursorTex, mousePos, 0, 1, r.White)

	r.EndDrawing()
}

// DrawUI renders the game UI
func (g *Game) DrawUI() {
	// Draw health bar
	if g.player != nil {
		barWidth := int32(300)
		barHeight := int32(20)
		barX := 400 - barWidth/2
		barY := int32(g.toolbarSlots[0].Y) - barHeight - 10

		r.DrawRectangle(barX, barY, barWidth, barHeight, r.DarkGray)
		healthWidth := int32(float32(g.player.CurrentHealth) / float32(g.player.MaxHealth) * float32(barWidth))
		r.DrawRectangle(barX, barY, healthWidth, barHeight, r.Red)
		r.DrawRectangleLinesEx(r.Rectangle{
			X:      float32(barX),
			Y:      float32(barY),
			Width:  float32(barWidth),
			Height: float32(barHeight),
		}, 2, r.Gray)

		healthText := fmt.Sprintf("%d/%d", g.player.CurrentHealth, g.player.MaxHealth)
		textWidth := r.MeasureText(healthText, 20)
		r.DrawText(healthText, barX+barWidth/2-textWidth/2, barY+2, 20, r.White)
	}

	// Draw timer
	minutes := int32(g.gameTimer) / 60
	seconds := int32(g.gameTimer) % 60
	timerText := fmt.Sprintf("%02d:%02d", minutes, seconds)
	textWidth := r.MeasureText(timerText, 30)
	r.DrawText(timerText, 400-textWidth/2, 10, 30, r.White)

	// Draw debug info
	g.DrawDebugInfo()

	// Draw inventory and crafting
	g.inventory.Draw(g.gameFont)
	g.crafting.Draw(g.gameFont, g.inventory)

	// Draw toolbar slots with weapons
	for i := range g.toolbarSlots {
		slotRect := g.toolbarSlots[i]

		// Draw base slot
		r.DrawRectangleRec(slotRect, r.Gray)

		// Only check weapon slots if player exists
		if g.player != nil && i < len(g.player.Weapons) {
			// Draw thicker border for current weapon slot
			if g.player.CurrentWeapon == g.player.Weapons[i] {
				thickness := float32(3)
				r.DrawRectangleLinesEx(slotRect, thickness, r.White)
			} else {
				r.DrawRectangleLinesEx(slotRect, 1, r.DarkGray)
			}

			// Draw weapon icons
			if weapon := g.player.Weapons[i]; weapon != nil {
				scale := float32(2.0)
				var tex r.Texture2D
				if raygun, ok := weapon.(*RayGun); ok {
					tex = raygun.Texture
				} else if sword, ok := weapon.(*Sword); ok {
					tex = sword.Texture
				} else if pistol, ok := weapon.(*Pistol); ok {
					tex = pistol.Texture
				}

				iconWidth := float32(tex.Width) * scale
				iconHeight := float32(tex.Height) * scale

				r.DrawTexturePro(
					tex,
					r.Rectangle{X: 0, Y: 0, Width: float32(tex.Width), Height: float32(tex.Height)},
					r.Rectangle{
						X:      slotRect.X + (slotRect.Width-iconWidth)/2,
						Y:      slotRect.Y + (slotRect.Height-iconHeight)/2,
						Width:  iconWidth,
						Height: iconHeight,
					},
					r.Vector2{X: 0, Y: 0},
					0,
					r.White,
				)
			}
		} else {
			// Draw normal border for empty slots
			r.DrawRectangleLinesEx(slotRect, 1, r.DarkGray)
		}
	}
}

// DrawDebugInfo renders debug information
func (g *Game) DrawDebugInfo() {
	if !g.debug {
		return
	}

	r.DrawTextEx(g.gameFont, fmt.Sprintf("FPS: %d", r.GetFPS()), r.Vector2{X: 10, Y: 10}, 20, 1, r.Green)
	if g.player != nil {
		r.DrawTextEx(g.gameFont, fmt.Sprintf("Player: %.0f, %.0f", g.player.X, g.player.Y), r.Vector2{X: 10, Y: 30}, 20, 1, r.Green)
		r.DrawTextEx(g.gameFont, fmt.Sprintf("Health: %d/%d", g.player.CurrentHealth, g.player.MaxHealth), r.Vector2{X: 10, Y: 50}, 20, 1, r.Green)
		r.DrawTextEx(g.gameFont, fmt.Sprintf("Exp: %d/%d", g.player.Experience, g.player.NextLevelExp), r.Vector2{X: 10, Y: 70}, 20, 1, r.Green)
	}
	r.DrawTextEx(g.gameFont, fmt.Sprintf("Camera: %.0f, %.0f", g.camera.Target.X, g.camera.Target.Y), r.Vector2{X: 10, Y: 90}, 20, 1, r.Green)
	r.DrawTextEx(g.gameFont, fmt.Sprintf("Trees: %d", len(g.trees)), r.Vector2{X: 10, Y: 110}, 20, 1, r.Green)
}

// Cleanup frees resources
func (g *Game) Cleanup() {
	for _, sprite := range g.sprites {
		sprite.Unload()
	}
	for _, tree := range g.trees {
		tree.Unload()
	}
	for _, item := range g.droppedItems {
		item.Unload()
	}
	if g.player != nil {
		g.player.Unload()
	}
	for _, enemy := range g.enemies {
		enemy.Unload()
	}
	for _, stone := range g.stones {
		stone.Unload()
	}
	r.UnloadTexture(g.cursorTex)
	r.ShowCursor()
	r.CloseWindow()
	g.inventory.Cleanup()
	if g.menu != nil {
		g.menu.Cleanup()
	}
	for _, portal := range g.portals {
		portal.Unload()
	}
	if g.merchant != nil {
		g.merchant.Unload()
	}
	if raygun, ok := g.player.CurrentWeapon.(*RayGun); ok {
		raygun.Unload()
	}
	for _, dummy := range g.dummies {
		dummy.Unload()
	}

	// Show the system cursor again
	r.ShowCursor()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	game := NewGame()
	game.Initialize()
	defer game.Cleanup()

	for !r.WindowShouldClose() {
		game.Update()
		game.Draw()
	}
}

func (g *Game) loadItemIcon(name, imagePath string) {
	if _, exists := g.inventory.ItemIcons[name]; !exists {
		g.inventory.ItemIcons[name] = r.LoadTexture(imagePath)
	}
}

// Add method to handle inventory item usage
func (g *Game) UpdateUI() {
	if g.inventory.IsOpen {
		if g.inventory.LastUsedItem == "Health Potion" {
			g.player.Heal(25)
			g.inventory.LastUsedItem = "" // Clear the last used item
		}
	}
}

// Add this helper function to check for object overlaps
func (g *Game) IsPositionOccupied(bounds r.Rectangle, padding float32) bool {
	// Check trees if they exist
	if g.trees != nil {
		for _, tree := range g.trees {
			if tree != nil { // Also check if individual tree is not nil
				treeBounds := r.Rectangle{
					X:      tree.X - padding,
					Y:      tree.Y - padding,
					Width:  float32(tree.Width) + padding*2,
					Height: float32(tree.Height) + padding*2,
				}
				if r.CheckCollisionRecs(bounds, treeBounds) {
					return true
				}
			}
		}
	}

	// Check stones if they exist
	if g.stones != nil {
		for _, stone := range g.stones {
			if stone != nil { // Also check if individual stone is not nil
				stoneBounds := r.Rectangle{
					X:      stone.X - padding,
					Y:      stone.Y - padding,
					Width:  float32(stone.Width) + padding*2,
					Height: float32(stone.Height) + padding*2,
				}
				if r.CheckCollisionRecs(bounds, stoneBounds) {
					return true
				}
			}
		}
	}

	// Check portals if they exist
	if g.portals != nil {
		for _, portal := range g.portals {
			if portal != nil { // Also check if individual portal is not nil
				portalBounds := r.Rectangle{
					X:      portal.X - padding,
					Y:      portal.Y - padding,
					Width:  float32(portal.Width) + padding*2,
					Height: float32(portal.Height) + padding*2,
				}
				if r.CheckCollisionRecs(bounds, portalBounds) {
					return true
				}
			}
		}
	}

	return false
}
