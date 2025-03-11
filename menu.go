package main

import (
	"fmt"

	r "github.com/gen2brain/raylib-go/raylib"
)

type MainMenu struct {
	playButton r.Rectangle
	bookIcon   struct {
		Texture  r.Texture2D
		Position r.Rectangle
		IsOpen   bool
	}
}

func (m *MainMenu) Initialize() {
	texture := r.LoadTexture("assets/book.png")
	if texture.ID == 0 {
		fmt.Println("Warning: Could not load book.png, using default texture")
		return
	}

	m.bookIcon.Texture = texture
	m.bookIcon.Position = r.Rectangle{
		X:      750,
		Y:      550,
		Width:  32,
		Height: 38,
	}
}

func NewMainMenu(_ r.Font) *MainMenu {
	return &MainMenu{
		playButton: r.Rectangle{
			X:      300,
			Y:      300,
			Width:  200,
			Height: 50,
		},
	}
}

func (m *MainMenu) Update() bool {
	if r.IsMouseButtonPressed(0) {
		mousePos := r.GetMousePosition()

		if m.bookIcon.IsOpen {
			m.bookIcon.IsOpen = false
			return false
		}

		if r.CheckCollisionPointRec(mousePos, m.playButton) {
			return true
		}

		if r.CheckCollisionPointRec(mousePos, m.bookIcon.Position) {
			m.bookIcon.IsOpen = true
		}
	}
	return false
}

func (m *MainMenu) Draw() {
	r.ClearBackground(r.Black)

	// Center KHAOZ title
	titleText := "KHAOZ"
	titleSize := int32(40)
	titleWidth := r.MeasureText(titleText, titleSize)
	titleX := 400 - titleWidth/2 // 400 is half of window width (800/2)
	r.DrawText(titleText, titleX, 200, titleSize, r.White)

	r.DrawRectangleRec(m.playButton, r.DarkGray)
	r.DrawText("PLAI", int32(m.playButton.X+70), int32(m.playButton.Y+10), 30, r.White)

	r.DrawText("v0.0 pre-alpha", 10, 570, 20, r.Gray)

	r.DrawTexturePro(
		m.bookIcon.Texture,
		r.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(m.bookIcon.Texture.Width),
			Height: float32(m.bookIcon.Texture.Height),
		},
		m.bookIcon.Position,
		r.Vector2{X: 0, Y: 0},
		0,
		r.White,
	)

	if m.bookIcon.IsOpen {
		r.DrawRectangle(150, 100, 500, 400, r.ColorAlpha(r.Black, 0.9))
		r.DrawRectangleLinesEx(r.Rectangle{X: 150, Y: 100, Width: 500, Height: 400}, 2, r.White)

		r.DrawText("Your lore thus far:", 170, 120, 20, r.White)

		loreText := "The entity sealed you away in pocket galaxies to stop you from saving humanity. Break the seals, power up and wreak havoc!"
		DrawTextBoxed(
			r.GetFontDefault(),
			loreText,
			r.Rectangle{X: 170, Y: 150, Width: 460, Height: 150},
			20,
			2,
			true,
			r.Yellow,
		)

		r.DrawText("Game Controls:", 170, 280, 20, r.White)
		r.DrawText("WASD - Move", 170, 310, 20, r.White)
		r.DrawText("E - Inventory", 170, 340, 20, r.White)
		r.DrawText("C - Crafting", 170, 370, 20, r.White)
		r.DrawText("Click - Interact", 170, 400, 20, r.White)
		r.DrawText("ESC - Close/Exit", 170, 430, 20, r.White)

		r.DrawText("Click anywhere to close", 270, 470, 20, r.Gray)
	}
}

func (m *MainMenu) Cleanup() {
	r.UnloadTexture(m.bookIcon.Texture)
}

func DrawTextBoxed(font r.Font, text string, rec r.Rectangle, fontSize float32, spacing float32, wordWrap bool, tint r.Color) {
	if len(text) == 0 {
		return
	}

	textOffsetY := float32(0)
	textOffsetX := float32(0)
	scaleFactor := fontSize / float32(font.BaseSize)
	words := []string{}
	currentWord := ""

	// Split text into words
	for i := 0; i < len(text); i++ {
		if text[i] == ' ' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
			words = append(words, " ")
		} else {
			currentWord += string(text[i])
		}
	}
	if currentWord != "" {
		words = append(words, currentWord)
	}

	// Draw words
	for _, word := range words {
		wordWidth := float32(r.MeasureText(word, int32(fontSize)))

		if wordWrap && (textOffsetX+wordWidth > rec.Width) {
			// Move to next line before drawing word
			textOffsetY += (float32(font.BaseSize) * float32(1.5)) * scaleFactor
			textOffsetX = 0
		}

		// Draw the word
		r.DrawTextEx(font, word, r.Vector2{
			X: rec.X + textOffsetX,
			Y: rec.Y + textOffsetY,
		}, fontSize, spacing, tint)

		textOffsetX += wordWidth + spacing
	}
}
