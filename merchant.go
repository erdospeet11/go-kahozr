package main

import (
	"fmt"

	r "github.com/gen2brain/raylib-go/raylib"
)

type ShopItem struct {
	Name     string
	Price    int
	IconPath string
}

type Merchant struct {
	X         float32
	Y         float32
	Width     int32
	Height    int32
	Texture   r.Texture2D
	IsOpen    bool
	ShopItems []ShopItem
	ItemIcons map[string]r.Texture2D
}

func NewMerchant(x, y float32) *Merchant {
	return &Merchant{
		X:         x,
		Y:         y,
		Width:     48,
		Height:    48,
		Texture:   r.LoadTexture("assets/merchant.png"),
		IsOpen:    false,
		ItemIcons: make(map[string]r.Texture2D),
		ShopItems: []ShopItem{
			{Name: "Health Potion", Price: 5, IconPath: "assets/health-potion.png"},
			{Name: "Strange Log", Price: 3, IconPath: "assets/tree-pickup.png"},
			{Name: "Stone Fragment", Price: 2, IconPath: "assets/stone-pickup.png"},
			{Name: "Golden Nugget", Price: 10, IconPath: "assets/gold-nugget.png"},
		},
	}
}

func (m *Merchant) LoadIcons() {
	for _, item := range m.ShopItems {
		if _, exists := m.ItemIcons[item.Name]; !exists {
			m.ItemIcons[item.Name] = r.LoadTexture(item.IconPath)
		}
	}
}

func (m *Merchant) Draw(gameFont r.Font, inventory *Inventory, camera r.Camera2D, debug bool) {
	// Draw merchant sprite in world space
	r.DrawTextureEx(
		m.Texture,
		r.Vector2{X: m.X, Y: m.Y},
		0,
		1,
		r.White,
	)

	// Draw interaction bounds in debug mode when shop is closed
	if debug && !m.IsOpen {
		interactionBounds := r.Rectangle{
			X:      m.X - 10,
			Y:      m.Y - 10,
			Width:  float32(m.Width) + 20,
			Height: float32(m.Height) + 20,
		}
		r.DrawRectangleLinesEx(interactionBounds, 1, r.Blue)
	}

	if !m.IsOpen {
		return
	}

	// Draw shop UI in screen space
	r.EndMode2D()

	// Draw shop window with increased width
	r.DrawRectangle(0, 0, 800, 600, r.ColorAlpha(r.Black, 0.5))
	r.DrawRectangle(200, 100, 400, 400, r.DarkGray) // Increased width from 300 to 400
	r.DrawTextEx(gameFont, "Merchant Shop", r.Vector2{X: 220, Y: 110}, 30, 1, r.White)

	// Draw items for sale
	y := 160
	iconSize := int32(32)
	for _, item := range m.ShopItems {
		if texture, exists := m.ItemIcons[item.Name]; exists {
			// Draw item icon
			r.DrawTexturePro(
				texture,
				r.Rectangle{X: 0, Y: 0, Width: float32(texture.Width), Height: float32(texture.Height)},
				r.Rectangle{X: 220, Y: float32(y), Width: float32(iconSize), Height: float32(iconSize)},
				r.Vector2{X: 0, Y: 0},
				0,
				r.White,
			)

			// Draw item name and price with adjusted positions
			r.DrawTextEx(gameFont,
				item.Name,
				r.Vector2{X: 220 + float32(iconSize) + 10, Y: float32(y)},
				20,
				1,
				r.White,
			)
			r.DrawTextEx(gameFont,
				fmt.Sprintf("%dg", item.Price),
				r.Vector2{X: 450, Y: float32(y)},
				20,
				1,
				r.Yellow,
			)

			// Draw buy button with adjusted position
			buyBtn := r.Rectangle{X: 520, Y: float32(y), Width: 50, Height: 25}
			canBuy := inventory.ItemCounts["Gold Coin"] >= item.Price

			if canBuy {
				r.DrawRectangleRec(buyBtn, r.Green)
			} else {
				r.DrawRectangleRec(buyBtn, r.Gray)
			}
			r.DrawTextEx(gameFont, "BUY", r.Vector2{X: buyBtn.X + 5, Y: buyBtn.Y + 2}, 20, 1, r.White)

			// Handle buy button click
			if canBuy && r.IsMouseButtonPressed(0) {
				mousePoint := r.GetMousePosition()
				if r.CheckCollisionPointRec(mousePoint, buyBtn) {
					m.BuyItem(item, inventory)
				}
			}
		}
		y += 40
	}

	// Draw close button with adjusted position
	closeBtn := r.Rectangle{X: 350, Y: 460, Width: 100, Height: 30}
	r.DrawRectangleRec(closeBtn, r.Gray)
	r.DrawTextEx(gameFont, "Close", r.Vector2{X: 370, Y: 465}, 20, 1, r.White)

	if r.IsMouseButtonPressed(0) {
		mousePoint := r.GetMousePosition()
		if r.CheckCollisionPointRec(mousePoint, closeBtn) {
			m.IsOpen = false
		}
	}

	r.BeginMode2D(camera)
}

func (m *Merchant) BuyItem(item ShopItem, inventory *Inventory) {
	// Remove gold coins from inventory
	inventory.ItemCounts["Gold Coin"] -= item.Price
	if inventory.ItemCounts["Gold Coin"] <= 0 {
		delete(inventory.ItemCounts, "Gold Coin")
	}

	// Add bought item to inventory
	inventory.Items = append(inventory.Items, item.Name)
	inventory.ItemCounts[item.Name]++
	inventory.LoadIcon(item.Name, item.IconPath)
}

func (m *Merchant) OnClick(mouseWorldPos r.Vector2) bool {
	return r.CheckCollisionPointRec(
		mouseWorldPos,
		r.Rectangle{
			X:      m.X - 10,
			Y:      m.Y - 10,
			Width:  float32(m.Width) + 20,
			Height: float32(m.Height) + 20,
		},
	)
}

func (m *Merchant) Unload() {
	r.UnloadTexture(m.Texture)
	for _, texture := range m.ItemIcons {
		r.UnloadTexture(texture)
	}
}

func (m *Merchant) GetBounds() r.Rectangle {
	return r.Rectangle{
		X:      m.X,
		Y:      m.Y,
		Width:  float32(m.Width),
		Height: float32(m.Height),
	}
}
