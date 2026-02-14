package gamepadutils

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"
)

// JoystickDisplay shows raw joystick buttons, axes, and hats.
type JoystickDisplay struct {
	renderer      *sdl.Renderer
	buttonTexture *sdl.Texture
	arrowTexture  *sdl.Texture
	buttonWidth   float32
	buttonHeight  float32
	arrowWidth    float32
	arrowHeight   float32

	Area sdl.FRect

	elementHighlighted string
	elementPressed     bool
}

// CreateJoystickDisplay creates a new joystick display widget.
func CreateJoystickDisplay(renderer *sdl.Renderer) *JoystickDisplay {
	ctx := &JoystickDisplay{
		renderer: renderer,
	}

	ctx.buttonTexture = CreateTextureFromPNG(renderer, gamepadButtonSmallPNG)
	if ctx.buttonTexture != nil {
		ctx.buttonWidth, ctx.buttonHeight, _ = ctx.buttonTexture.Size()
	}

	ctx.arrowTexture = CreateTextureFromPNG(renderer, gamepadAxisArrowPNG)
	if ctx.arrowTexture != nil {
		ctx.arrowWidth, ctx.arrowHeight, _ = ctx.arrowTexture.Size()
	}

	return ctx
}

// SetArea sets the display area.
func (ctx *JoystickDisplay) SetArea(area *sdl.FRect) {
	if ctx == nil {
		return
	}
	ctx.Area = *area
}

// GetElementAt returns the joystick element string at the given coordinates.
func (ctx *JoystickDisplay) GetElementAt(joystick *sdl.Joystick, x, y float32) string {
	if ctx == nil {
		return ""
	}

	nbuttons, _ := joystick.NumButtons()
	naxes, _ := joystick.NumAxes()
	nhats, _ := joystick.NumHats()

	const margin = 8.0
	const center = 80.0
	const arrowExtent = 48.0

	point := sdl.FPoint{X: x, Y: y}
	cx := ctx.Area.X + margin
	cy := ctx.Area.Y + margin

	if nbuttons > 0 {
		cy += FONT_LINE_HEIGHT + 2

		for i := int32(0); i < nbuttons; i++ {
			highlight := sdl.FRect{
				X: cx,
				Y: cy + FONT_CHARACTER_SIZE/2 - ctx.buttonHeight/2,
				W: center - margin*2,
				H: ctx.buttonHeight,
			}
			if PointInRectFloat(&point, &highlight) {
				return fmt.Sprintf("b%d", i)
			}
			cy += ctx.buttonHeight + 2
		}
	}

	cx = ctx.Area.X + margin + center + margin
	cy = ctx.Area.Y + margin

	if naxes > 0 {
		cy += FONT_LINE_HEIGHT + 2

		for i := int32(0); i < naxes; i++ {
			text := fmt.Sprintf("%d:", i)

			highlight := sdl.FRect{
				X: cx + FONT_CHARACTER_SIZE*float32(len(text)) + 2.0,
				Y: cy + FONT_CHARACTER_SIZE/2 - ctx.buttonHeight/2,
				W: ctx.arrowWidth + arrowExtent,
				H: ctx.buttonHeight,
			}
			if PointInRectFloat(&point, &highlight) {
				return fmt.Sprintf("-a%d", i)
			}

			highlight.X += highlight.W
			if PointInRectFloat(&point, &highlight) {
				return fmt.Sprintf("+a%d", i)
			}

			cy += ctx.buttonHeight + 2
		}
	}

	cy += FONT_LINE_HEIGHT + 2

	if nhats > 0 {
		cy += FONT_LINE_HEIGHT + 2 + 1.5*ctx.buttonHeight - FONT_CHARACTER_SIZE/2

		for i := int32(0); i < nhats; i++ {
			text := fmt.Sprintf("%d:", i)

			dst := sdl.FRect{
				X: cx + FONT_CHARACTER_SIZE*float32(len(text)) + 2,
				Y: cy + FONT_CHARACTER_SIZE/2 - ctx.buttonHeight/2,
				W: ctx.buttonWidth,
				H: ctx.buttonHeight,
			}
			if PointInRectFloat(&point, &dst) {
				return fmt.Sprintf("h%d.%d", i, sdl.HAT_LEFT)
			}

			dst.X += ctx.buttonWidth
			dst.Y -= ctx.buttonHeight
			if PointInRectFloat(&point, &dst) {
				return fmt.Sprintf("h%d.%d", i, sdl.HAT_UP)
			}

			dst.Y += ctx.buttonHeight * 2
			if PointInRectFloat(&point, &dst) {
				return fmt.Sprintf("h%d.%d", i, sdl.HAT_DOWN)
			}

			dst.X += ctx.buttonWidth
			dst.Y = cy + FONT_CHARACTER_SIZE/2 - ctx.buttonHeight/2
			if PointInRectFloat(&point, &dst) {
				return fmt.Sprintf("h%d.%d", i, sdl.HAT_RIGHT)
			}

			cy += 3*ctx.buttonHeight + 2
		}
	}
	return ""
}

// SetHighlight sets the highlighted element.
func (ctx *JoystickDisplay) SetHighlight(element string, pressed bool) {
	if ctx == nil {
		return
	}
	ctx.elementHighlighted = element
	ctx.elementPressed = pressed
}

// renderButtonHighlight highlights a joystick button.
func (ctx *JoystickDisplay) renderButtonHighlight(button int, area *sdl.FRect) {
	if ctx.elementHighlighted == "" || ctx.elementHighlighted[0] != 'b' {
		return
	}
	val, err := strconv.Atoi(ctx.elementHighlighted[1:])
	if err != nil || val != button {
		return
	}
	saved := saveColor(ctx.renderer)
	if ctx.elementPressed {
		ctx.renderer.SetDrawColor(PRESSED_R, PRESSED_G, PRESSED_B, PRESSED_A)
	} else {
		ctx.renderer.SetDrawColor(HIGHLIGHT_R, HIGHLIGHT_G, HIGHLIGHT_B, HIGHLIGHT_A)
	}
	ctx.renderer.RenderFillRect(area)
	restoreColor(ctx.renderer, saved)
}

// renderAxisHighlight highlights a joystick axis direction.
func (ctx *JoystickDisplay) renderAxisHighlight(axis int, direction int, area *sdl.FRect) {
	prefix := byte('+')
	if direction < 0 {
		prefix = '-'
	}
	if ctx.elementHighlighted == "" ||
		ctx.elementHighlighted[0] != prefix ||
		len(ctx.elementHighlighted) < 2 ||
		ctx.elementHighlighted[1] != 'a' {
		return
	}
	val, err := strconv.Atoi(ctx.elementHighlighted[2:])
	if err != nil || val != axis {
		return
	}
	saved := saveColor(ctx.renderer)
	if ctx.elementPressed {
		ctx.renderer.SetDrawColor(PRESSED_R, PRESSED_G, PRESSED_B, PRESSED_A)
	} else {
		ctx.renderer.SetDrawColor(HIGHLIGHT_R, HIGHLIGHT_G, HIGHLIGHT_B, HIGHLIGHT_A)
	}
	ctx.renderer.RenderFillRect(area)
	restoreColor(ctx.renderer, saved)
}

// setupHatHighlight checks if a hat direction should be highlighted and applies the texture color mod.
func (ctx *JoystickDisplay) setupHatHighlight(hat int, direction uint8) bool {
	if ctx.elementHighlighted == "" || ctx.elementHighlighted[0] != 'h' {
		return false
	}
	// Parse "h<hat>.<dir>"
	dotIdx := strings.IndexByte(ctx.elementHighlighted, '.')
	if dotIdx < 0 || dotIdx < 2 {
		return false
	}
	hatVal, err1 := strconv.Atoi(ctx.elementHighlighted[1:dotIdx])
	dirVal, err2 := strconv.Atoi(ctx.elementHighlighted[dotIdx+1:])
	if err1 != nil || err2 != nil {
		return false
	}
	if hatVal == hat && uint8(dirVal) == direction {
		if ctx.elementPressed {
			ctx.buttonTexture.SetColorMod(PRESSED_TEXTURE_MOD[0], PRESSED_TEXTURE_MOD[1], PRESSED_TEXTURE_MOD[2])
		} else {
			ctx.buttonTexture.SetColorMod(HIGHLIGHT_TEXTURE_MOD[0], HIGHLIGHT_TEXTURE_MOD[1], HIGHLIGHT_TEXTURE_MOD[2])
		}
		return true
	}
	return false
}

// Render draws the joystick display.
func (ctx *JoystickDisplay) Render(joystick *sdl.Joystick) {
	if ctx == nil {
		return
	}

	nbuttons, _ := joystick.NumButtons()
	naxes, _ := joystick.NumAxes()
	nhats, _ := joystick.NumHats()

	const margin = 8.0
	const center = 80.0
	const arrowExtent = 48.0

	saved := saveColor(ctx.renderer)

	x := ctx.Area.X + margin
	y := ctx.Area.Y + margin

	if nbuttons > 0 {
		ctx.renderer.DebugText(x, y, "BUTTONS")
		y += FONT_LINE_HEIGHT + 2

		for i := int32(0); i < nbuttons; i++ {
			highlight := sdl.FRect{
				X: x,
				Y: y + FONT_CHARACTER_SIZE/2 - ctx.buttonHeight/2,
				W: center - margin*2,
				H: ctx.buttonHeight,
			}
			ctx.renderButtonHighlight(int(i), &highlight)

			text := fmt.Sprintf("%2d:", i)
			ctx.renderer.DebugText(x, y, text)

			if joystick.Button(i) {
				ctx.buttonTexture.SetColorMod(10, 255, 21)
			} else {
				ctx.buttonTexture.SetColorMod(255, 255, 255)
			}

			dst := sdl.FRect{
				X: x + FONT_CHARACTER_SIZE*float32(len(text)) + 2,
				Y: y + FONT_CHARACTER_SIZE/2 - ctx.buttonHeight/2,
				W: ctx.buttonWidth,
				H: ctx.buttonHeight,
			}
			ctx.renderer.RenderTexture(ctx.buttonTexture, nil, &dst)

			y += ctx.buttonHeight + 2
		}
	}

	x = ctx.Area.X + margin + center + margin
	y = ctx.Area.Y + margin

	if naxes > 0 {
		ctx.renderer.DebugText(x, y, "AXES")
		y += FONT_LINE_HEIGHT + 2

		for i := int32(0); i < naxes; i++ {
			value, _ := joystick.Axis(i)

			text := fmt.Sprintf("%d:", i)
			ctx.renderer.DebugText(x, y, text)

			highlight := sdl.FRect{
				X: x + FONT_CHARACTER_SIZE*float32(len(text)) + 2.0,
				Y: y + FONT_CHARACTER_SIZE/2 - ctx.buttonHeight/2,
				W: ctx.arrowWidth + arrowExtent,
				H: ctx.buttonHeight,
			}
			ctx.renderAxisHighlight(int(i), -1, &highlight)

			highlight.X += highlight.W
			ctx.renderAxisHighlight(int(i), 1, &highlight)

			dst := sdl.FRect{
				X: x + FONT_CHARACTER_SIZE*float32(len(text)) + 2.0,
				Y: y + FONT_CHARACTER_SIZE/2 - ctx.arrowHeight/2,
				W: ctx.arrowWidth,
				H: ctx.arrowHeight,
			}

			if value == math.MinInt16 {
				ctx.arrowTexture.SetColorMod(10, 255, 21)
			} else {
				ctx.arrowTexture.SetColorMod(255, 255, 255)
			}
			ctx.renderer.RenderTextureRotated(ctx.arrowTexture, nil, &dst, 0.0, nil, sdl.FLIP_HORIZONTAL)

			dst.X += ctx.arrowWidth

			ctx.renderer.SetDrawColor(200, 200, 200, 255)
			rect := sdl.FRect{
				X: dst.X + arrowExtent - 2.0,
				Y: dst.Y,
				W: 4.0,
				H: ctx.arrowHeight,
			}
			ctx.renderer.RenderFillRect(&rect)
			restoreColor(ctx.renderer, saved)

			if value < 0 {
				ctx.renderer.SetDrawColor(8, 200, 16, 255)
				rect.W = (float32(value) / float32(math.MinInt16)) * arrowExtent
				rect.X = dst.X + arrowExtent - rect.W
				rect.Y = dst.Y + ctx.arrowHeight*0.25
				rect.H = ctx.arrowHeight / 2.0
				ctx.renderer.RenderFillRect(&rect)
			}

			dst.X += arrowExtent

			if value > 0 {
				ctx.renderer.SetDrawColor(8, 200, 16, 255)
				rect.W = (float32(value) / float32(math.MaxInt16)) * arrowExtent
				rect.X = dst.X
				rect.Y = dst.Y + ctx.arrowHeight*0.25
				rect.H = ctx.arrowHeight / 2.0
				ctx.renderer.RenderFillRect(&rect)
			}

			dst.X += arrowExtent

			if value == math.MaxInt16 {
				ctx.arrowTexture.SetColorMod(10, 255, 21)
			} else {
				ctx.arrowTexture.SetColorMod(255, 255, 255)
			}
			ctx.renderer.RenderTexture(ctx.arrowTexture, nil, &dst)

			restoreColor(ctx.renderer, saved)

			y += ctx.buttonHeight + 2
		}
	}

	y += FONT_LINE_HEIGHT + 2

	if nhats > 0 {
		ctx.renderer.DebugText(x, y, "HATS")
		y += FONT_LINE_HEIGHT + 2 + 1.5*ctx.buttonHeight - FONT_CHARACTER_SIZE/2

		for i := int32(0); i < nhats; i++ {
			value := joystick.Hat(i)

			text := fmt.Sprintf("%d:", i)
			ctx.renderer.DebugText(x, y, text)

			// Left
			if value&sdl.HAT_LEFT != 0 {
				ctx.buttonTexture.SetColorMod(10, 255, 21)
			} else if !ctx.setupHatHighlight(int(i), sdl.HAT_LEFT) {
				ctx.buttonTexture.SetColorMod(255, 255, 255)
			}
			dst := sdl.FRect{
				X: x + FONT_CHARACTER_SIZE*float32(len(text)) + 2,
				Y: y + FONT_CHARACTER_SIZE/2 - ctx.buttonHeight/2,
				W: ctx.buttonWidth,
				H: ctx.buttonHeight,
			}
			ctx.renderer.RenderTexture(ctx.buttonTexture, nil, &dst)

			// Up
			if value&sdl.HAT_UP != 0 {
				ctx.buttonTexture.SetColorMod(10, 255, 21)
			} else if !ctx.setupHatHighlight(int(i), sdl.HAT_UP) {
				ctx.buttonTexture.SetColorMod(255, 255, 255)
			}
			dst.X += ctx.buttonWidth
			dst.Y -= ctx.buttonHeight
			ctx.renderer.RenderTexture(ctx.buttonTexture, nil, &dst)

			// Down
			if value&sdl.HAT_DOWN != 0 {
				ctx.buttonTexture.SetColorMod(10, 255, 21)
			} else if !ctx.setupHatHighlight(int(i), sdl.HAT_DOWN) {
				ctx.buttonTexture.SetColorMod(255, 255, 255)
			}
			dst.Y += ctx.buttonHeight * 2
			ctx.renderer.RenderTexture(ctx.buttonTexture, nil, &dst)

			// Right
			if value&sdl.HAT_RIGHT != 0 {
				ctx.buttonTexture.SetColorMod(10, 255, 21)
			} else if !ctx.setupHatHighlight(int(i), sdl.HAT_RIGHT) {
				ctx.buttonTexture.SetColorMod(255, 255, 255)
			}
			dst.X += ctx.buttonWidth
			dst.Y = y + FONT_CHARACTER_SIZE/2 - ctx.buttonHeight/2
			ctx.renderer.RenderTexture(ctx.buttonTexture, nil, &dst)

			y += 3*ctx.buttonHeight + 2
		}
	}
}

// Destroy releases the joystick display resources.
func (ctx *JoystickDisplay) Destroy() {
	if ctx == nil {
		return
	}
	if ctx.buttonTexture != nil {
		ctx.buttonTexture.Destroy()
	}
	if ctx.arrowTexture != nil {
		ctx.arrowTexture.Destroy()
	}
}
