package wioterminal

import (
	"testing"
	"time"
)

func TestButtonDebounce(t *testing.T) {
	var b buttonState
	start := time.Unix(0, 0)
	steps := []struct {
		ms   int
		down bool
		want bool
	}{
		{0, true, false},
		{5, false, false},
		{10, true, false},
		{29, true, false},
		{30, true, true},
		{100, true, false}, // Holding does not repeat.
		{105, false, false},
		{110, true, false}, // Release chatter does not rearm.
		{140, true, false},
		{150, false, false},
		{170, false, false},
		{180, true, false},
		{200, true, true},
	}
	for _, step := range steps {
		if got := b.update(start.Add(time.Duration(step.ms)*time.Millisecond), step.down); got != step.want {
			t.Fatalf("at %d ms: got %v, want %v", step.ms, got, step.want)
		}
	}
}

func TestInputState(t *testing.T) {
	var state inputState
	now := time.Unix(0, 0)
	// Indices into buttons: Up Up Down Down Left Right Left Right B A.
	command := []int{0, 0, 1, 1, 2, 3, 2, 3, 4, 5}
	sample := func(down [len(buttons)]bool) bool {
		t.Helper()
		now = now.Add(10 * time.Millisecond)
		if state.update(now, down) {
			t.Fatal("accepted an unsettled input")
		}
		now = now.Add(debounceDelay)
		return state.update(now, down)
	}
	press := func(index int) bool {
		t.Helper()
		var down [len(buttons)]bool
		down[index] = true
		got := sample(down)
		now = now.Add(time.Second)
		if state.update(now, down) {
			t.Fatal("holding a button repeated a command")
		}
		if sample([len(buttons)]bool{}) {
			t.Fatal("release completed a command")
		}
		return got
	}
	for attempt := 0; attempt < 2; attempt++ {
		for i, index := range command {
			if got, want := press(index), i == len(command)-1; got != want {
				t.Fatalf("attempt %d, input %d: got %v, want %v", attempt, i, got, want)
			}
		}
	}
	// A chord must interrupt a partial command, including when one key is
	// pressed before the other instead of both settling in the same poll.
	for _, staggered := range []bool{false, true} {
		for _, index := range command[:8] {
			press(index)
		}
		var down [len(buttons)]bool
		if staggered {
			down[4] = true
			sample(down)
		}
		down[4], down[5] = true, true
		if sample(down) {
			t.Fatal("simultaneous B/A completed a command")
		}
		sample([len(buttons)]bool{})
		if press(4) || press(5) {
			t.Fatal("partial command survived simultaneous presses")
		}
	}
}
