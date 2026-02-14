package gamepadutils

import (
	"bytes"
	"embed"
	"image"
	"image/color"
	_ "image/png"
	"unsafe"

	"github.com/Zyko0/go-sdl3/sdl"
)

//go:embed assets/gamepad_front.png
var gamepadFrontPNG []byte

//go:embed assets/gamepad_back.png
var gamepadBackPNG []byte

//go:embed assets/gamepad_face_abxy.png
var gamepadFaceABXYPNG []byte

//go:embed assets/gamepad_face_axby.png
var gamepadFaceAXBYPNG []byte

//go:embed assets/gamepad_face_bayx.png
var gamepadFaceBAYXPNG []byte

//go:embed assets/gamepad_face_sony.png
var gamepadFaceSonyPNG []byte

//go:embed assets/gamepad_battery.png
var gamepadBatteryPNG []byte

//go:embed assets/gamepad_battery_wired.png
var gamepadBatteryWiredPNG []byte

//go:embed assets/gamepad_touchpad.png
var gamepadTouchpadPNG []byte

//go:embed assets/gamepad_button.png
var gamepadButtonPNG []byte

//go:embed assets/gamepad_button_small.png
var gamepadButtonSmallPNG []byte

//go:embed assets/gamepad_axis.png
var gamepadAxisPNG []byte

//go:embed assets/gamepad_axis_arrow.png
var gamepadAxisArrowPNG []byte

//go:embed assets/gamepad_button_background.png
var gamepadButtonBackgroundPNG []byte

//go:embed assets/gamepad_wired.png
var gamepadWiredPNG []byte

//go:embed assets/gamepad_wireless.png
var gamepadWirelessPNG []byte

//go:embed assets/gamecontrollerdb.txt
var GamecontrollerDB []byte

// Embed the assets directory
//
//go:embed assets
var _ embed.FS

// CreateTextureFromPNG loads a PNG from embedded bytes and creates an SDL texture.
func CreateTextureFromPNG(renderer *sdl.Renderer, data []byte) *sdl.Texture {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	// Convert to RGBA
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	surface, err := sdl.CreateSurfaceFrom(w, h, sdl.PIXELFORMAT_RGBA32,
		unsafe.Slice((*byte)(unsafe.Pointer(&rgba.Pix[0])), len(rgba.Pix)),
		rgba.Stride)
	if err != nil {
		return nil
	}
	defer surface.Destroy()

	// Handle grayscale images - set colorkey for transparency
	// The front/back images use grayscale with black as transparent
	if _, ok := img.(*image.Gray); ok {
		// For grayscale images, convert with alpha channel
		rgbaImg := image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c := img.At(x, y)
				r, g, b, _ := c.RGBA()
				if r == 0 && g == 0 && b == 0 {
					rgbaImg.Set(x, y, color.RGBA{0, 0, 0, 0})
				} else {
					rgbaImg.Set(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255})
				}
			}
		}
		surface2, err := sdl.CreateSurfaceFrom(w, h, sdl.PIXELFORMAT_RGBA32,
			unsafe.Slice((*byte)(unsafe.Pointer(&rgbaImg.Pix[0])), len(rgbaImg.Pix)),
			rgbaImg.Stride)
		if err != nil {
			return nil
		}
		defer surface2.Destroy()
		texture, err := renderer.CreateTextureFromSurface(surface2)
		if err != nil {
			return nil
		}
		return texture
	}

	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil
	}
	return texture
}
