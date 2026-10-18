package main

import (
	"image/color"
	"testing"
)

func TestRainbowColor(t *testing.T) {
	want := []color.RGBA{
		{255, 0, 0, 255}, {255, 255, 0, 255}, {0, 255, 0, 255},
		{0, 255, 255, 255}, {0, 0, 255, 255}, {255, 0, 255, 255},
	}
	for i, c := range want {
		if got := rainbowColor(uint16(i * 256)); got != c {
			t.Errorf("phase %d: got %v, want %v", i*256, got, c)
		}
	}
	for phase := uint16(0); phase < 1536; phase++ {
		if rainbowColor(phase) != rainbowColor(phase+1536) {
			t.Fatalf("phase %d does not repeat", phase)
		}
		a, b := rainbowColor(phase), rainbowColor(phase+1)
		for i, channel := range [][2]uint8{{a.R, b.R}, {a.G, b.G}, {a.B, b.B}} {
			diff := int(channel[0]) - int(channel[1])
			if diff < -1 || diff > 1 {
				t.Fatalf("phase %d: channel %d jumps by %d", phase, i, diff)
			}
		}
	}
}
