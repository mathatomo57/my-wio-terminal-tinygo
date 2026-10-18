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

func main() {
	for _, p := range pages {
		p.button.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}

	display := initdisplay.InitDisplay()
	display.FillScreen(black)
	drawPNG(display, pages[0].png)

	pressed := make([]bool, len(pages))
	for {
		for i, p := range pages {
			down := !p.button.Get()
			if down && !pressed[i] {
				drawPNG(display, p.png)
			}
			pressed[i] = down
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func drawPNG(display *ili9341.Device, b []byte) {
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
