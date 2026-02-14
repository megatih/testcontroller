package gamepadutils

import (
	"fmt"

	"github.com/Zyko0/go-sdl3/sdl"
)

// GamepadTypeDisplay shows a list of gamepad types for selection in binding mode.
type GamepadTypeDisplay struct {
	renderer *sdl.Renderer

	typeHighlighted int
	typePressed     bool
	typeSelected    int
	realType        sdl.GamepadType

	Area sdl.FRect
}

// CreateGamepadTypeDisplay creates a new type display widget.
func CreateGamepadTypeDisplay(renderer *sdl.Renderer) *GamepadTypeDisplay {
	return &GamepadTypeDisplay{
		renderer:        renderer,
		typeHighlighted: SDL_GAMEPAD_TYPE_UNSELECTED,
		typeSelected:    SDL_GAMEPAD_TYPE_UNSELECTED,
		realType:        sdl.GAMEPAD_TYPE_UNKNOWN,
	}
}

// SetArea sets the display area.
func (ctx *GamepadTypeDisplay) SetArea(area *sdl.FRect) {
	if ctx == nil {
		return
	}
	ctx.Area = *area
}

// SetHighlight sets the highlighted type.
func (ctx *GamepadTypeDisplay) SetHighlight(t int, pressed bool) {
	if ctx == nil {
		return
	}
	ctx.typeHighlighted = t
	ctx.typePressed = pressed
}

// SetSelected sets the selected type.
func (ctx *GamepadTypeDisplay) SetSelected(t int) {
	if ctx == nil {
		return
	}
	ctx.typeSelected = t
}

// SetRealType sets the actual gamepad type.
func (ctx *GamepadTypeDisplay) SetRealType(t sdl.GamepadType) {
	if ctx == nil {
		return
	}
	ctx.realType = t
}

// GetTypeAt returns the type at the given screen coordinates.
func (ctx *GamepadTypeDisplay) GetTypeAt(x, y float32) int {
	if ctx == nil {
		return SDL_GAMEPAD_TYPE_UNSELECTED
	}

	const margin = 8.0
	const lineHeight = 16.0

	point := sdl.FPoint{X: x, Y: y}
	cy := ctx.Area.Y + margin

	for i := int(sdl.GAMEPAD_TYPE_UNKNOWN); i < int(sdl.GAMEPAD_TYPE_COUNT); i++ {
		highlight := sdl.FRect{
			X: ctx.Area.X + margin,
			Y: cy,
			W: ctx.Area.W - margin*2,
			H: lineHeight,
		}
		if PointInRectFloat(&point, &highlight) {
			return i
		}
		cy += lineHeight
	}
	return SDL_GAMEPAD_TYPE_UNSELECTED
}

// renderTypeHighlight draws a highlight for a gamepad type.
func (ctx *GamepadTypeDisplay) renderTypeHighlight(t int, area *sdl.FRect) {
	if t == ctx.typeHighlighted || t == ctx.typeSelected {
		saved := saveColor(ctx.renderer)

		if t == ctx.typeHighlighted {
			if ctx.typePressed {
				ctx.renderer.SetDrawColor(PRESSED_R, PRESSED_G, PRESSED_B, PRESSED_A)
			} else {
				ctx.renderer.SetDrawColor(HIGHLIGHT_R, HIGHLIGHT_G, HIGHLIGHT_B, HIGHLIGHT_A)
			}
		} else {
			ctx.renderer.SetDrawColor(SELECTED_R, SELECTED_G, SELECTED_B, SELECTED_A)
		}
		ctx.renderer.RenderFillRect(area)
		restoreColor(ctx.renderer, saved)
	}
}

// Render draws the type display.
func (ctx *GamepadTypeDisplay) Render() {
	if ctx == nil {
		return
	}

	const margin = 8.0
	const lineHeight = 16.0

	x := ctx.Area.X + margin
	y := ctx.Area.Y + margin

	for i := int(sdl.GAMEPAD_TYPE_UNKNOWN); i < int(sdl.GAMEPAD_TYPE_COUNT); i++ {
		highlight := sdl.FRect{
			X: x,
			Y: y,
			W: ctx.Area.W - margin*2,
			H: lineHeight,
		}
		ctx.renderTypeHighlight(i, &highlight)

		var text string
		if i == int(sdl.GAMEPAD_TYPE_UNKNOWN) {
			if ctx.realType == sdl.GAMEPAD_TYPE_UNKNOWN ||
				ctx.realType == sdl.GAMEPAD_TYPE_STANDARD {
				text = "Auto (Standard)"
			} else {
				text = fmt.Sprintf("Auto (%s)", GetGamepadTypeString(ctx.realType))
			}
		} else if i == int(sdl.GAMEPAD_TYPE_STANDARD) {
			text = "Standard"
		} else {
			text = GetGamepadTypeString(sdl.GamepadType(i))
		}

		ctx.renderer.DebugText(x+margin, y+lineHeight/2-FONT_CHARACTER_SIZE/2, text)
		y += lineHeight
	}
}

// Destroy releases resources (no textures to free).
func (ctx *GamepadTypeDisplay) Destroy() {
}
