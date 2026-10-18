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

var sequence = [...]Button{Up, Up, Down, Down, Left, Right, Left, Right, B, A}

// prefix holds the length of the longest proper prefix that is also a suffix
// at each position in sequence. It allows overlapping attempts to be retained.
var prefix = [...]uint8{0, 1, 0, 0, 0, 0, 0, 0, 0, 0}

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
	for d.matched > 0 && button != sequence[d.matched] {
		d.matched = prefix[d.matched-1]
	}
	if button == sequence[d.matched] {
		d.matched++
	}
	if int(d.matched) == len(sequence) {
		d.matched = prefix[d.matched-1]
		return true
	}
	return false
}

// Reset discards a partially entered command.
func (d *Detector) Reset() {
	d.matched = 0
}
