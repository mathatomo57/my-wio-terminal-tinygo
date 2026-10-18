package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/examples/ili9341/initdisplay"
	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/drivers/image/png"
	"tinygo.org/x/drivers/lis3dh"
)

//go:embed x_icon_bg.png
var xIconBgPNG []byte

// 144x144 RGB565 (big endian)。角は背景と同じ赤。
//
//go:embed x_icon.rgb565
var xIcon []byte

//go:embed x_qr.png
var xQRPNG []byte

//go:embed github_qr.png
var githubQRPNG []byte

// 上面のボタン (画面に向かって左から C, B, A) と表示する画像の対応。
// 先頭のページだけ加速度センサーに応じてアイコンが動く。
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
	// アイコンが動ける範囲 (下は名前の文字を避ける)
	iconMaxX = 320 - iconSize
	iconMaxY = 200 - iconSize
	// この傾き (µg) で端まで動く
	tiltFull = 500_000
	// これ未満の移動は無視する (センサーのノイズ対策)
	moveMin = 3
)

var (
	black = color.RGBA{0, 0, 0, 255}
	red   = color.RGBA{227, 51, 36, 255} // x_icon_bg.png の背景と同じ値
)

// png デコーダは 1 行ごとに callback を呼ぶので、画面幅分あれば足りる。
var buffer [320]uint16

var display *ili9341.Device

func main() {
	for _, p := range pages {
		p.button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}

	machine.I2C0.Configure(machine.I2CConfig{SCL: machine.SCL0_PIN, SDA: machine.SDA0_PIN})
	accel := lis3dh.New(machine.I2C0)
	if err := accel.Configure(lis3dh.Config{Address: lis3dh.Address0}); err != nil {
		fatal(err)
	}
	if err := accel.SetRange(lis3dh.RANGE_2_G); err != nil {
		fatal(err)
	}

	display = initdisplay.InitDisplay()
	display.FillScreen(black)

	current := 0
	drawPNG(pages[current].png)
	icon := newIcon(iconMaxX/2, iconMaxY/2)

	pressed := make([]bool, len(pages))
	var fx, fy int32 // 平滑化した加速度
	for {
		for i, p := range pages {
			down := !p.button.Get()
			if down && !pressed[i] && i != current {
				current = i
				drawPNG(p.png)
				if current == 0 {
					icon = newIcon(iconMaxX/2, iconMaxY/2)
				}
			}
			pressed[i] = down
		}

		if current == 0 {
			ax, ay, _, _ := accel.ReadAcceleration()
			fx += (ax - fx) / 4
			fy += (ay - fy) / 4
			// センサーの Y 軸が画面の左右、X 軸が画面の上下に対応する
			icon.moveTo(
				clamp(iconMaxX/2-int(int64(fy)*(iconMaxX/2)/tiltFull), 0, iconMaxX),
				clamp(iconMaxY/2+int(int64(fx)*(iconMaxY/2)/tiltFull), 0, iconMaxY),
			)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

type icon struct{ x, y int }

func newIcon(x, y int) icon {
	display.DrawRGBBitmap8(int16(x), int16(y), xIcon, iconSize, iconSize)
	return icon{x, y}
}

// 前の位置からはみ出した帯を赤で消してから、新しい位置に描く。
func (ic *icon) moveTo(x, y int) {
	if abs(x-ic.x) < moveMin && abs(y-ic.y) < moveMin {
		return
	}
	if x > ic.x {
		display.FillRectangle(int16(ic.x), int16(ic.y), int16(x-ic.x), iconSize, red)
	} else if x < ic.x {
		display.FillRectangle(int16(x+iconSize), int16(ic.y), int16(ic.x-x), iconSize, red)
	}
	if y > ic.y {
		display.FillRectangle(int16(ic.x), int16(ic.y), iconSize, int16(y-ic.y), red)
	} else if y < ic.y {
		display.FillRectangle(int16(ic.x), int16(y+iconSize), iconSize, int16(ic.y-y), red)
	}
	display.DrawRGBBitmap8(int16(x), int16(y), xIcon, iconSize, iconSize)
	ic.x, ic.y = x, y
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

func drawPNG(b []byte) {
	png.SetCallback(buffer[:], func(data []uint16, x, y, w, h, width, height int16) {
		if err := display.DrawRGBBitmap(x, y, data[:w*h], w, h); err != nil {
			fatal(err)
		}
	})
	if _, err := png.Decode(bytes.NewReader(b)); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	for {
		fmt.Printf("%s\r\n", err.Error())
		time.Sleep(5 * time.Second)
	}
}
