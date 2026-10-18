package main

import "image/color"

// rainbowColor は 1536 段階で一周する色相を RGB に変換する。
func rainbowColor(phase uint16) color.RGBA {
	phase %= 1536
	rise := uint8(phase % 256)
	fall := 255 - rise
	switch phase / 256 {
	case 0:
		return color.RGBA{255, rise, 0, 255}
	case 1:
		return color.RGBA{fall, 255, 0, 255}
	case 2:
		return color.RGBA{0, 255, rise, 255}
	case 3:
		return color.RGBA{0, fall, 255, 255}
	case 4:
		return color.RGBA{rise, 0, 255, 255}
	default:
		return color.RGBA{255, 0, fall, 255}
	}
}
