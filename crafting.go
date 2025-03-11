package main

import (
	"fmt"
	"sort"

	r "github.com/gen2brain/raylib-go/raylib"
)

// Recipe represents a crafting recipe
type Recipe struct {
	Result     string
	ResultIcon r.Texture2D
	Materials  map[string]int
}

// CraftingSystem represents the crafting interface
type CraftingSystem struct {
	IsOpen  bool
	Recipes []Recipe
}

// NewCraftingSystem creates a new crafting system
func NewCraftingSystem() *CraftingSystem {
	return &CraftingSystem{
		IsOpen: false,
		Recipes: []Recipe{
			{
				Result: "Pickaxe",
				Materials: map[string]int{
					"Strange Log":    2,
					"Stone Fragment": 3,
				},
			},
			{
				Result: "Gold Coin",
				Materials: map[string]int{
					"Golden Nugget": 3,
				},
			},
		},
	}
}

// Draw renders the crafting UI
func (cs *CraftingSystem) Draw(gameFont r.Font, inventory *Inventory) {
	if !cs.IsOpen {
		return
	}

	// Draw semi-transparent background
	r.DrawRectangle(0, 0, 800, 600, r.ColorAlpha(r.Black, 0.5))

	// Draw crafting panel
	r.DrawRectangle(150, 100, 500, 400, r.DarkGray)
	r.DrawTextEx(gameFont, "Crafting", r.Vector2{X: 170, Y: 110}, 30, 1, r.White)

	// Draw recipes
	y := 160
	iconSize := int32(20)
	for _, recipe := range cs.Recipes {
		// Draw result item icon and name
		if texture, exists := inventory.ItemIcons[recipe.Result]; exists {
			r.DrawTexturePro(
				texture,
				r.Rectangle{X: 0, Y: 0, Width: float32(texture.Width), Height: float32(texture.Height)},
				r.Rectangle{X: 170, Y: float32(y), Width: float32(iconSize), Height: float32(iconSize)},
				r.Vector2{X: 0, Y: 0},
				0,
				r.White,
			)
		}
		r.DrawTextEx(gameFont, recipe.Result, r.Vector2{X: 170 + float32(iconSize) + 5, Y: float32(y)}, 20, 1, r.White)

		// Get sorted material names
		var materialNames []string
		for item := range recipe.Materials {
			materialNames = append(materialNames, item)
		}
		sort.Strings(materialNames)

		// Draw required materials
		x := float32(300)
		for _, item := range materialNames {
			count := recipe.Materials[item]
			if texture, exists := inventory.ItemIcons[item]; exists {
				r.DrawTexturePro(
					texture,
					r.Rectangle{X: 0, Y: 0, Width: float32(texture.Width), Height: float32(texture.Height)},
					r.Rectangle{X: x, Y: float32(y), Width: float32(iconSize), Height: float32(iconSize)},
					r.Vector2{X: 0, Y: 0},
					0,
					r.White,
				)
				r.DrawTextEx(gameFont, fmt.Sprintf("x%d", count), r.Vector2{X: x + float32(iconSize) + 2, Y: float32(y)}, 20, 1, r.White)
				x += float32(iconSize) + 40
			}
		}

		// Draw craft button
		craftBtn := r.Rectangle{X: 550, Y: float32(y), Width: 70, Height: 20}
		canCraft := cs.CanCraft(recipe, inventory)

		if canCraft {
			r.DrawRectangleRec(craftBtn, r.Green)
		} else {
			r.DrawRectangleRec(craftBtn, r.Gray)
		}
		r.DrawTextEx(gameFont, "CRAFT", r.Vector2{X: craftBtn.X + 5, Y: craftBtn.Y}, 20, 1, r.White)

		// Handle craft button click
		if canCraft && r.IsMouseButtonPressed(0) {
			mousePoint := r.GetMousePosition()
			if r.CheckCollisionPointRec(mousePoint, craftBtn) {
				cs.CraftItem(recipe, inventory)
			}
		}

		y += 40
	}

	// Draw close button
	closeBtn := r.Rectangle{X: 350, Y: 460, Width: 100, Height: 30}
	r.DrawRectangleRec(closeBtn, r.Gray)
	r.DrawTextEx(gameFont, "Close", r.Vector2{X: 370, Y: 465}, 20, 1, r.White)

	if r.IsMouseButtonPressed(0) {
		mousePoint := r.GetMousePosition()
		if r.CheckCollisionPointRec(mousePoint, closeBtn) {
			cs.IsOpen = false
		}
	}
}

// CanCraft checks if a recipe can be crafted
func (cs *CraftingSystem) CanCraft(recipe Recipe, inventory *Inventory) bool {
	for item, needed := range recipe.Materials {
		if inventory.ItemCounts[item] < needed {
			return false
		}
	}
	return true
}

// CraftItem performs the crafting operation
func (cs *CraftingSystem) CraftItem(recipe Recipe, inventory *Inventory) {
	// Remove materials
	for item, count := range recipe.Materials {
		inventory.ItemCounts[item] -= count
		if inventory.ItemCounts[item] <= 0 {
			delete(inventory.ItemCounts, item)
		}
	}
	// Add crafted item
	inventory.Items = append(inventory.Items, recipe.Result)
	inventory.ItemCounts[recipe.Result]++
}
