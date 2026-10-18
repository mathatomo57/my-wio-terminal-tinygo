// Package cheatcode detects the sequence Up, Up, Down, Down, Left, Right,
// Left, Right, B, A without depending on any hardware drivers.
package cheatcode

// Button identifies a logical button. Map hardware events to these values.
type Button uint8

const (
	Up Button = iota + 1
	Down
	Left
	Right
	B
	A
)

// commandButtons is the fixed command recognized by this package.
var commandButtons = [...]Button{Up, Up, Down, Down, Left, Right, Left, Right, B, A}

// fallbackMatchCounts[n-1] is how many of the n matched buttons can be reused
// as the start of a new attempt when the next button does not match.
// After Up, Up, the last Up can be reused as the first Up, so entry 1 is 1.
// No other matched portion ends with a reusable start of this command.
// This table must be updated together with commandButtons.
var fallbackMatchCounts = [...]uint8{0, 1, 0, 0, 0, 0, 0, 0, 0, 0}

// Detector tracks progress through the command. Its zero value is ready to use.
// A Detector must not be used concurrently without external synchronization.
type Detector struct {
	matched uint8
}

// Input consumes one button press and reports whether it completes the command.
// Call Input once per press, rather than repeatedly while a button is held.
// Unknown buttons interrupt the current attempt. After completion, the detector
// is ready for another command.
func (d *Detector) Input(button Button) bool {
	for d.matched > 0 && button != commandButtons[d.matched] {
		d.matched = fallbackMatchCounts[d.matched-1]
	}
	if button == commandButtons[d.matched] {
		d.matched++
	}
	if int(d.matched) == len(commandButtons) {
		d.matched = fallbackMatchCounts[d.matched-1]
		return true
	}
	return false
}

// Reset discards a partially entered command.
func (d *Detector) Reset() {
	d.matched = 0
}
