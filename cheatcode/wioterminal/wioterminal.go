//go:build wioterminal

package wioterminal

import (
	"machine"
	"time"
)

// pins has the same order as buttons in state.go.
var pins = [...]machine.Pin{
	machine.WIO_5S_UP, machine.WIO_5S_DOWN, machine.WIO_5S_LEFT,
	machine.WIO_5S_RIGHT, machine.WIO_KEY_B, machine.WIO_KEY_A,
}

// Detector reads and debounces the board's directional switch and A/B buttons.
// Create it with New and use it from a single goroutine.
type Detector struct {
	state inputState
}

// New configures the six input pins with pull-ups. Use one detector per board.
func New() *Detector {
	for _, pin := range pins {
		pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}
	return &Detector{}
}

// Poll reads the controls without sleeping and reports command completion once.
// Call it regularly, for example every 5–10 ms. A press and its release must
// each settle for at least 20 ms. Held buttons do not repeat. Simultaneous
// presses interrupt the current command; C and the center switch are ignored.
func (d *Detector) Poll() bool {
	var down [len(buttons)]bool
	for i, pin := range pins {
		down[i] = !pin.Get()
	}
	return d.state.update(time.Now(), down)
}

// Reset discards command progress while preserving debounce and held-key state.
func (d *Detector) Reset() {
	d.state.command.Reset()
}
