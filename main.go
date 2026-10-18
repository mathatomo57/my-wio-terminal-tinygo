package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"machine"
	"runtime"
	"time"

	"github.com/masatomo57/my-wio-terminal-tinygo/cheatcode/wioterminal"
	"tinygo.org/x/drivers/examples/ili9341/initdisplay"
	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/drivers/image/png"
	"tinygo.org/x/drivers/lis3dh"
)

//go:embed x_icon_bg.png
var xIconBgPNG []byte

// 144x144 RGB565 (big endian)。
//
//go:embed x_icon.rgb565
var xIcon []byte

//go:embed x_qr.png
var xQRPNG []byte

//go:embed github_qr.png
var githubQRPNG []byte

// 上面のボタン (画面に向かって左から C, B, A) と表示する画像の対応。
var pages = []struct {
	button machine.Pin
	png    []byte
}{
	{machine.WIO_KEY_C, xIconBgPNG},
	{machine.WIO_KEY_B, xQRPNG},
	{machine.WIO_KEY_A, githubQRPNG},
}

const (
	iconSize = 144
	iconMaxX = 320 - iconSize
	iconMaxY = 200 - iconSize
	tiltFull = 500_000
	moveMin  = 3
)

var black = color.RGBA{0, 0, 0, 255}
var red = color.RGBA{227, 51, 36, 255}

// png デコーダは 1 行ごとに callback を呼ぶので、画面幅分あれば足りる。
var buffer [320]uint16

// 背景の位置だけを保持し、アニメーション中の PNG 再デコードを避ける。
var background [240][5]uint64

func main() {
	detector := wioterminal.New()
	for _, p := range pages {
		p.button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}

	if err := machine.I2C0.Configure(machine.I2CConfig{SCL: machine.SCL0_PIN, SDA: machine.SDA0_PIN}); err != nil {
		fatal(err)
	}
	accel := lis3dh.New(machine.I2C0)
	if err := accel.Configure(lis3dh.Config{Address: lis3dh.Address0}); err != nil {
		fatal(err)
	}
	if err := accel.SetRange(lis3dh.RANGE_2_G); err != nil {
		fatal(err)
	}
	display := initdisplay.InitDisplay()
	display.FillScreen(black)
	drawPNG(display, pages[0].png)
	ic := movingIcon{iconMaxX / 2, iconMaxY / 2}
	ic.draw(display, false, 0)

	// 画像のデコードや描画中も、押下・解放を一定間隔で観測する。
	commands := make(chan struct{}, 1)
	go func() {
		for {
			if detector.Poll() {
				commands <- struct{}{}
			}
			time.Sleep(5 * time.Millisecond)
		}
	}()

	pressed := make([]bool, len(pages))
	currentPage := 0
	rainbow := false
	var phase uint16
	var lastFrame time.Time
	var lastTilt time.Time
	var fx, fy int32
	for {
		select {
		case <-commands:
			rainbow = !rainbow
			if !rainbow {
				drawPNG(display, pages[currentPage].png)
				if currentPage == 0 {
					ic.draw(display, false, phase)
				}
			}
		default:
		}
		for i, p := range pages {
			down := !p.button.Get()
			if down && !pressed[i] && i != currentPage {
				currentPage = i
				drawPNG(display, p.png)
				if currentPage == 0 {
					ic = movingIcon{iconMaxX / 2, iconMaxY / 2}
					ic.draw(display, rainbow, phase)
				}
				lastFrame = time.Time{}
			}
			pressed[i] = down
		}
		if currentPage == 0 && time.Since(lastTilt) >= 20*time.Millisecond {
			ax, ay, _, err := accel.ReadAcceleration()
			if err != nil {
				fatal(err)
			}
			fx += (ax - fx) / 4
			fy += (ay - fy) / 4
			// センサーの Y 軸が画面の左右、X 軸が画面の上下に対応する。
			x := clamp(iconMaxX/2-int(int64(fy)*(iconMaxX/2)/tiltFull), 0, iconMaxX)
			y := clamp(iconMaxY/2+int(int64(fx)*(iconMaxY/2)/tiltFull), 0, iconMaxY)
			if abs(x-ic.x) >= moveMin || abs(y-ic.y) >= moveMin {
				ic.erase(display, rainbow, phase)
				ic.x, ic.y = x, y
				ic.draw(display, rainbow, phase)
			}
			lastTilt = time.Now()
		}
		if rainbow && time.Since(lastFrame) >= 50*time.Millisecond {
			phase = (phase + 6) % 1536
			var sprite *movingIcon
			if currentPage == 0 {
				sprite = &ic
			}
			drawRainbow(display, phase, sprite)
			if currentPage == 0 {
				ic.draw(display, true, phase)
			}
			lastFrame = time.Now()
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func drawPNG(display *ili9341.Device, b []byte) {
	background = [240][5]uint64{}
	var backgroundColor uint16
	png.SetCallback(buffer[:], func(data []uint16, x, y, w, h, width, height int16) {
		// 画像の左上の色を背景色とする。同色の画素だけを置き換える。
		if x == 0 && y == 0 {
			backgroundColor = data[0]
		}
		for i, pixel := range data[:w*h] {
			px, py := int(x)+i%int(w), int(y)+i/int(w)
			if pixel == backgroundColor && px < 320 && py < 240 {
				background[py][px/64] |= uint64(1) << uint(px%64)
			}
		}
		if err := display.DrawRGBBitmap(x, y, data[:w*h], w, h); err != nil {
			fatal(err)
		}
		runtime.Gosched()
	})
	if _, err := png.Decode(bytes.NewReader(b)); err != nil {
		fatal(err)
	}
}

func drawRainbow(display *ili9341.Device, phase uint16, sprite *movingIcon) {
	for y := 0; y < 240; y++ {
		c := rainbowColor((phase + uint16(y)*4) % 1536)
		for x := 0; x < 320; {
			if !backgroundAt(x, y, sprite) {
				x++
				continue
			}
			start := x
			for x < 320 && backgroundAt(x, y, sprite) {
				x++
			}
			if err := display.FillRectangle(int16(start), int16(y), int16(x-start), 1, c); err != nil {
				fatal(err)
			}
		}
		runtime.Gosched()
	}
}

func backgroundAt(x, y int, sprite *movingIcon) bool {
	// アイコン領域は draw で背景ごと描くため、先に塗りつぶさない。
	if sprite != nil && x >= sprite.x && x < sprite.x+iconSize && y >= sprite.y && y < sprite.y+iconSize {
		return false
	}
	return background[y][x/64]&(uint64(1)<<uint(x%64)) != 0
}

func fatal(err error) {
	for {
		fmt.Printf("%s\r\n", err.Error())
		time.Sleep(5 * time.Second)
	}
}

// アイコンの移動領域は名前の文字と重ならず、下地は背景色のみ。
type movingIcon struct{ x, y int }

func (ic movingIcon) erase(display *ili9341.Device, rainbow bool, phase uint16) {
	for y := ic.y; y < ic.y+iconSize; y++ {
		c := red
		if rainbow {
			c = rainbowColor((phase + uint16(y)*4) % 1536)
		}
		if err := display.FillRectangle(int16(ic.x), int16(y), iconSize, 1, c); err != nil {
			fatal(err)
		}
		runtime.Gosched()
	}
}

func (ic movingIcon) draw(display *ili9341.Device, rainbow bool, phase uint16) {
	backgroundColor := ili9341.RGBATo565(red)
	for y := 0; y < iconSize; y++ {
		replacement := backgroundColor
		if rainbow {
			replacement = ili9341.RGBATo565(rainbowColor((phase + uint16(ic.y+y)*4) % 1536))
		}
		for x := 0; x < iconSize; x++ {
			offset := (y*iconSize + x) * 2
			pixel := uint16(xIcon[offset])<<8 | uint16(xIcon[offset+1])
			if pixel == backgroundColor {
				pixel = replacement
			}
			buffer[x] = pixel
		}
		if err := display.DrawRGBBitmap(int16(ic.x), int16(ic.y+y), buffer[:iconSize], iconSize, 1); err != nil {
			fatal(err)
		}
		runtime.Gosched()
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
