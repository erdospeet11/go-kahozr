package main

import (
	"fmt"
	"math/rand"
	"sort"

	r "github.com/gen2brain/raylib-go/raylib"
)

// Inventory represents the player's inventory
type Inventory struct {
	IsOpen       bool
	Items        []string
	ItemCounts   map[string]int
	ItemIcons    map[string]r.Texture2D
	LastUsedItem string
	player       *Player
}

// NewInventory creates a new inventory instance
func NewInventory(player *Player) *Inventory {
	return &Inventory{
		IsOpen:     false,
		Items:      make([]string, 0),
		ItemCounts: make(map[string]int),
		ItemIcons:  make(map[string]r.Texture2D),
		player:     player,
	}
}

// Draw renders the inventory UI
func (inv *Inventory) Draw(gameFont r.Font) {
	if !inv.IsOpen {
		return
	}

	// Draw semi-transparent background
	r.DrawRectangle(0, 0, 800, 600, r.ColorAlpha(r.Black, 0.5))

	// Draw wider inventory panel
	r.DrawRectangle(50, 100, 700, 400, r.DarkGray)

	// Draw left column header (Items)
	r.DrawTextEx(gameFont, "Inventory", r.Vector2{X: 70, Y: 110}, 30, 1, r.White)

	// Draw right column header (Weapons)
	r.DrawTextEx(gameFont, "Weapons", r.Vector2{X: 550, Y: 110}, 30, 1, r.White)

	// Update left column X positions
	leftX := float32(70)
	// Draw items in left column
	y := 160
	iconSize := int32(20)
	for _, item := range inv.GetSortedItems() {
		if count, exists := inv.ItemCounts[item]; exists && count > 0 {
			if icon, hasIcon := inv.ItemIcons[item]; hasIcon {
				r.DrawTexturePro(
					icon,
					r.Rectangle{X: 0, Y: 0, Width: float32(icon.Width), Height: float32(icon.Height)},
					r.Rectangle{X: leftX, Y: float32(y), Width: float32(iconSize), Height: float32(iconSize)},
					r.Vector2{X: 0, Y: 0},
					0,
					r.White,
				)
			}
			r.DrawTextEx(gameFont, fmt.Sprintf("%s x%d", item, count), r.Vector2{X: leftX + float32(iconSize) + 5, Y: float32(y)}, 20, 1, r.White)

			// Add USE button for usable items
			if item == "Goodie Bag" || item == "Health Potion" {
				useBtn := r.Rectangle{X: leftX + 350, Y: float32(y), Width: 50, Height: 25}
				r.DrawRectangleRec(useBtn, r.Green)

				// Center the USE text
				textWidth := r.MeasureTextEx(gameFont, "USE", 20, 1).X
				textX := useBtn.X + (useBtn.Width-textWidth)/2
				textY := useBtn.Y + (useBtn.Height-20)/2 // 20 is the font size
				r.DrawTextEx(gameFont, "USE", r.Vector2{X: textX, Y: textY}, 20, 1, r.White)

				// Handle click
				if r.IsMouseButtonPressed(0) {
					mousePoint := r.GetMousePosition()
					if r.CheckCollisionPointRec(mousePoint, useBtn) {
						inv.UseItem(item)
					}
				}
			}
			y += 40
		}
	}

	// Update right column X positions
	rightX := float32(550)
	// Draw weapons in right column
	y = 160
	for _, weapon := range inv.player.Weapons {
		var tex r.Texture2D
		var name string

		switch w := weapon.(type) {
		case *RayGun:
			tex = w.Texture
			name = "Ray Gun"
		case *Sword:
			tex = w.Texture
			name = "Sword"
		case *Pistol:
			tex = w.Texture
			name = "Pistol"
		}

		// Draw weapon icon
		r.DrawTexturePro(
			tex,
			r.Rectangle{X: 0, Y: 0, Width: float32(tex.Width), Height: float32(tex.Height)},
			r.Rectangle{X: rightX, Y: float32(y), Width: float32(iconSize), Height: float32(iconSize)},
			r.Vector2{X: 0, Y: 0},
			0,
			r.White,
		)

		// Draw weapon name
		r.DrawTextEx(gameFont, name, r.Vector2{X: rightX + float32(iconSize) + 5, Y: float32(y)}, 20, 1, r.White)

		y += 40
	}

	// Draw close button with click handling
	closeBtn := r.Rectangle{X: 350, Y: 450, Width: 100, Height: 30}
	r.DrawRectangleRec(closeBtn, r.Gray)
	r.DrawTextEx(gameFont, "Close", r.Vector2{X: 375, Y: 455}, 20, 1, r.White)

	// Check close button click
	if r.IsMouseButtonPressed(0) {
		mousePoint := r.GetMousePosition()
		if r.CheckCollisionPointRec(mousePoint, closeBtn) {
			inv.IsOpen = false
		}
	}
}

// LoadIcon loads an item icon if it doesn't exist
func (inv *Inventory) LoadIcon(name, imagePath string) {
	if _, exists := inv.ItemIcons[name]; !exists {
		inv.ItemIcons[name] = r.LoadTexture(imagePath)
	}
}

// Cleanup frees resources
func (inv *Inventory) Cleanup() {
	for _, texture := range inv.ItemIcons {
		r.UnloadTexture(texture)
	}
}

// Update UseItem method to handle both Goodie Bag and Health Potion
func (inv *Inventory) UseItem(itemName string) {
	if inv.ItemCounts[itemName] > 0 {
		inv.LastUsedItem = itemName
		switch itemName {
		case "Goodie Bag":
			// Remove one goodie bag
			inv.ItemCounts[itemName]--
			if inv.ItemCounts[itemName] <= 0 {
				delete(inv.ItemCounts, itemName)
				// Remove from items list
				for j, item := range inv.Items {
					if item == itemName {
						inv.Items = append(inv.Items[:j], inv.Items[j+1:]...)
						break
					}
				}
			}

			// Add random loot
			lootTable := []string{
				"Stone Fragment",
				"Strange Log",
				"Golden Nugget",
				"Pickaxe",
			}
			randomLoot := lootTable[rand.Intn(len(lootTable))]
			inv.Items = append(inv.Items, randomLoot)
			inv.ItemCounts[randomLoot]++
			inv.LoadIcon(randomLoot, getIconPath(randomLoot))
			inv.LastUsedItem = randomLoot

		case "Health Potion":
			// Remove one health potion
			inv.ItemCounts[itemName]--
			if inv.ItemCounts[itemName] <= 0 {
				delete(inv.ItemCounts, itemName)
			}
			// Heal player (this will be handled in game.go)
		}
	}
}

// Add this helper function to get the correct icon path:
func getIconPath(itemName string) string {
	switch itemName {
	case "Health Potion":
		return "assets/health-potion.png"
	case "Stone Fragment":
		return "assets/stone-pickup.png"
	case "Strange Log":
		return "assets/tree-pickup.png"
	case "Golden Nugget":
		return "assets/gold-nugget.png"
	case "Pickaxe":
		return "assets/pickaxe.png"
	default:
		return "assets/" + itemName + ".png"
	}
}

// Add this method to Inventory struct
func (inv *Inventory) GetSortedItems() []string {
	var items []string
	for item := range inv.ItemCounts {
		items = append(items, item)
	}
	sort.Strings(items)
	return items
}

// Add Pickaxe to the item icons
func (inv *Inventory) LoadDefaultIcons() {
	inv.LoadIcon("Pickaxe", "assets/pickaxe.png")
	// ... other icons ...
}
