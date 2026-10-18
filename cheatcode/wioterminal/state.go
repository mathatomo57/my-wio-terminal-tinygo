// Package wioterminal reads the Wio Terminal controls to detect cheat codes.
// New, Detector, and its hardware methods are available with the wioterminal
// build tag supplied by TinyGo's -target wioterminal option.
package wioterminal

import (
	"time"

	"github.com/masatomo57/my-wio-terminal-tinygo/cheatcode"
)

const debounceDelay = 20 * time.Millisecond

var buttons = [...]cheatcode.Button{
	cheatcode.Up, cheatcode.Down, cheatcode.Left,
	cheatcode.Right, cheatcode.B, cheatcode.A,
}

type buttonState struct {
	raw, stable bool
	changed     time.Time
}

// update reports a press only after the reading has settled. Releases are also
// debounced, so release chatter cannot create a second press.
func (b *buttonState) update(now time.Time, down bool) bool {
	if down != b.raw {
		b.raw = down
		b.changed = now
	}
	if b.raw == b.stable || now.Sub(b.changed) < debounceDelay {
		return false
	}
	b.stable = b.raw
	return b.stable
}

type inputState struct {
	command cheatcode.Detector
	keys    [len(buttons)]buttonState
}

func (s *inputState) update(now time.Time, down [len(buttons)]bool) bool {
	pressed := -1
	held := 0
	for i := range s.keys {
		if s.keys[i].update(now, down[i]) {
			pressed = i
		}
		if s.keys[i].stable {
			held++
		}
	}
	// There is no defined order for simultaneous presses. Discard the attempt
	// rather than inventing an order from the pin iteration order.
	if held > 1 {
		s.command.Reset()
		return false
	}
	return pressed >= 0 && s.command.Input(buttons[pressed])
}
