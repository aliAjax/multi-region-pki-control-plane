package clock

import (
	"testing"
	"time"
)

func TestFixedClockPreservesLifecycleBoundary(t *testing.T) {
	w := time.Unix(100, 0)
	if g := (Fixed{T: w}).Now(); !g.Equal(w) {
		t.Fatalf("drift: %v", g)
	}
}

func TestFixedClockAdvanceUsesConfiguredTime(t *testing.T) {
	start := time.Unix(100, 0).UTC()
	got := (Fixed{T: start}).Advance(2 * time.Hour).Now()
	if !got.Equal(start.Add(2 * time.Hour)) {
		t.Fatalf("advance drifted: %v", got)
	}
}
