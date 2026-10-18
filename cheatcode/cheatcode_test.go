package cheatcode_test

import (
	"testing"

	"github.com/mathatomo57/my-wio-terminal-tinygo/cheatcode"
)

func TestInput(t *testing.T) {
	command := []cheatcode.Button{
		cheatcode.Up, cheatcode.Up, cheatcode.Down, cheatcode.Down,
		cheatcode.Left, cheatcode.Right, cheatcode.Left, cheatcode.Right,
		cheatcode.B, cheatcode.A,
	}
	tests := []struct {
		name   string
		prefix []cheatcode.Button
	}{
		{name: "complete command"},
		{name: "unrelated input", prefix: []cheatcode.Button{cheatcode.A, cheatcode.B}},
		{name: "overlapping start", prefix: []cheatcode.Button{cheatcode.Up}},
		{name: "restart after partial command", prefix: command[:6]},
		{name: "unknown button interrupts", prefix: []cheatcode.Button{cheatcode.Up, cheatcode.Up, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d cheatcode.Detector
			for _, button := range tt.prefix {
				if d.Input(button) {
					t.Fatal("completed during prefix")
				}
			}
			// The same detector must also recognize a second consecutive command.
			for attempt := 0; attempt < 2; attempt++ {
				for i, button := range command {
					if got, want := d.Input(button), i == len(command)-1; got != want {
						t.Fatalf("attempt %d, input %d: got %v, want %v", attempt, i, got, want)
					}
				}
			}
		})
	}
}

func TestReset(t *testing.T) {
	var d cheatcode.Detector
	d.Input(cheatcode.Up)
	d.Input(cheatcode.Up)
	d.Reset()
	for _, button := range []cheatcode.Button{
		cheatcode.Down, cheatcode.Down, cheatcode.Left, cheatcode.Right,
		cheatcode.Left, cheatcode.Right, cheatcode.B, cheatcode.A,
	} {
		if d.Input(button) {
			t.Fatal("completed a command using input from before Reset")
		}
	}
}
