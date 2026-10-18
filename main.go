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
)

//go:embed x_icon.png
var xIconPNG []byte

//go:embed x_qr.png
var xQRPNG []byte

//go:embed github_qr.png
var githubQRPNG []byte

// 上面のボタン (画面に向かって左から C, B, A) と表示する画像の対応。
var pages = []struct {
	button machine.Pin
	png    []byte
}{
	{machine.WIO_KEY_C, xIconPNG},
	{machine.WIO_KEY_B, xQRPNG},
	{machine.WIO_KEY_A, githubQRPNG},
}

var black = color.RGBA{0, 0, 0, 255}

// png デコーダは 1 行ごとに callback を呼ぶので、画面幅分あれば足りる。
var buffer [320]uint16

// 背景の位置だけを保持し、アニメーション中の PNG 再デコードを避ける。
var background [240][5]uint64

func main() {
	detector := wioterminal.New()
	for _, p := range pages {
		p.button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}

	display := initdisplay.InitDisplay()
	display.FillScreen(black)
	drawPNG(display, pages[0].png)

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
	for {
		select {
		case <-commands:
			rainbow = !rainbow
			if !rainbow {
				drawPNG(display, pages[currentPage].png)
			}
		default:
		}
		for i, p := range pages {
			down := !p.button.Get()
			if down && !pressed[i] {
				currentPage = i
				drawPNG(display, p.png)
				lastFrame = time.Time{}
			}
			pressed[i] = down
		}
		if rainbow && time.Since(lastFrame) >= 50*time.Millisecond {
			drawRainbow(display, phase)
			phase = (phase + 6) % 1536
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

func drawRainbow(display *ili9341.Device, phase uint16) {
	for y := 0; y < 240; y++ {
		c := rainbowColor((phase + uint16(y)*4) % 1536)
		for x := 0; x < 320; {
			if background[y][x/64]&(uint64(1)<<uint(x%64)) == 0 {
				x++
				continue
			}
			start := x
			for x < 320 && background[y][x/64]&(uint64(1)<<uint(x%64)) != 0 {
				x++
			}
			if err := display.FillRectangle(int16(start), int16(y), int16(x-start), 1, c); err != nil {
				fatal(err)
			}
		}
		runtime.Gosched()
	}
}

func fatal(err error) {
	for {
		fmt.Printf("%s\r\n", err.Error())
		time.Sleep(5 * time.Second)
	}
}
