package gamepadutils

import "github.com/Zyko0/go-sdl3/sdl"

// GamepadButton is a UI button widget with 9-slice background rendering.
type GamepadButton struct {
	renderer         *sdl.Renderer
	background       *sdl.Texture
	backgroundWidth  float32
	backgroundHeight float32

	area sdl.FRect

	label       string
	labelWidth  float32
	labelHeight float32

	highlight bool
	pressed   bool
}

// CreateGamepadButton creates a new button widget.
func CreateGamepadButton(renderer *sdl.Renderer, label string) *GamepadButton {
	ctx := &GamepadButton{
		renderer: renderer,
	}

	ctx.background = CreateTextureFromPNG(renderer, gamepadButtonBackgroundPNG)
	if ctx.background != nil {
		w, h, _ := ctx.background.Size()
		ctx.backgroundWidth = w
		ctx.backgroundHeight = h
	}

	ctx.SetLabel(label)
	return ctx
}

// SetLabel sets the button's label text.
func (b *GamepadButton) SetLabel(label string) {
	if b == nil {
		return
	}
	b.label = label
	b.labelWidth = FONT_CHARACTER_SIZE * float32(len(label))
	b.labelHeight = FONT_CHARACTER_SIZE
}

// SetArea sets the button's bounding rectangle.
func (b *GamepadButton) SetArea(area *sdl.FRect) {
	if b == nil {
		return
	}
	b.area = *area
}

// GetArea returns the button's bounding rectangle.
func (b *GamepadButton) GetArea() sdl.FRect {
	if b == nil {
		return sdl.FRect{}
	}
	return b.area
}

// SetHighlight sets the highlight and pressed state.
func (b *GamepadButton) SetHighlight(highlight, pressed bool) {
	if b == nil {
		return
	}
	b.highlight = highlight
	if highlight {
		b.pressed = pressed
	} else {
		b.pressed = false
	}
}

// LabelWidth returns the width of the label text.
func (b *GamepadButton) LabelWidth() float32 {
	if b == nil {
		return 0
	}
	return b.labelWidth
}

// LabelHeight returns the height of the label text.
func (b *GamepadButton) LabelHeight() float32 {
	if b == nil {
		return 0
	}
	return b.labelHeight
}

// Contains returns true if the point is inside the button area.
func (b *GamepadButton) Contains(x, y float32) bool {
	if b == nil {
		return false
	}
	point := sdl.FPoint{X: x, Y: y}
	return PointInRectFloat(&point, &b.area)
}

// Render draws the button widget using 9-slice rendering.
func (b *GamepadButton) Render() {
	if b == nil || b.background == nil {
		return
	}

	oneThirdW := b.backgroundWidth / 3
	oneThirdH := b.backgroundHeight / 3

	if b.pressed {
		b.background.SetColorMod(PRESSED_TEXTURE_MOD[0], PRESSED_TEXTURE_MOD[1], PRESSED_TEXTURE_MOD[2])
	} else if b.highlight {
		b.background.SetColorMod(HIGHLIGHT_TEXTURE_MOD[0], HIGHLIGHT_TEXTURE_MOD[1], HIGHLIGHT_TEXTURE_MOD[2])
	} else {
		b.background.SetColorMod(255, 255, 255)
	}

	// Top left
	src := sdl.FRect{X: 0, Y: 0, W: oneThirdW, H: oneThirdH}
	dst := sdl.FRect{X: b.area.X, Y: b.area.Y, W: src.W, H: src.H}
	b.renderer.RenderTexture(b.background, &src, &dst)

	// Bottom left
	src.Y = b.backgroundHeight - oneThirdH
	dst.Y = b.area.Y + b.area.H - oneThirdH
	b.renderer.RenderTexture(b.background, &src, &dst)

	// Bottom right
	src.X = b.backgroundWidth - oneThirdW
	dst.X = b.area.X + b.area.W - oneThirdW
	b.renderer.RenderTexture(b.background, &src, &dst)

	// Top right
	src.Y = 0
	dst.Y = b.area.Y
	b.renderer.RenderTexture(b.background, &src, &dst)

	// Left edge
	src = sdl.FRect{X: 0, Y: oneThirdH, W: oneThirdW, H: b.backgroundHeight - 2*oneThirdH}
	dst = sdl.FRect{X: b.area.X, Y: b.area.Y + oneThirdH, W: oneThirdW, H: b.area.H - 2*oneThirdH}
	b.renderer.RenderTexture(b.background, &src, &dst)

	// Right edge
	src.X = b.backgroundWidth - oneThirdW
	dst.X = b.area.X + b.area.W - oneThirdW
	b.renderer.RenderTexture(b.background, &src, &dst)

	// Top edge
	src = sdl.FRect{X: oneThirdW, Y: 0, W: b.backgroundWidth - 2*oneThirdW, H: oneThirdH}
	dst = sdl.FRect{X: b.area.X + oneThirdW, Y: b.area.Y, W: b.area.W - 2*oneThirdW, H: oneThirdH}
	b.renderer.RenderTexture(b.background, &src, &dst)

	// Bottom edge
	src.Y = b.backgroundHeight - oneThirdH
	dst.Y = b.area.Y + b.area.H - oneThirdH
	b.renderer.RenderTexture(b.background, &src, &dst)

	// Center
	src = sdl.FRect{X: oneThirdW, Y: oneThirdH, W: b.backgroundWidth - 2*oneThirdW, H: b.backgroundHeight - 2*oneThirdH}
	dst = sdl.FRect{X: b.area.X + oneThirdW, Y: b.area.Y + oneThirdH, W: b.area.W - 2*oneThirdW, H: b.area.H - 2*oneThirdH}
	b.renderer.RenderTexture(b.background, &src, &dst)

	// Label centered
	labelX := b.area.X + b.area.W/2 - b.labelWidth/2
	labelY := b.area.Y + b.area.H/2 - b.labelHeight/2
	b.renderer.DebugText(labelX, labelY, b.label)
}

// Destroy releases the button's resources.
func (b *GamepadButton) Destroy() {
	if b == nil {
		return
	}
	if b.background != nil {
		b.background.Destroy()
	}
}
