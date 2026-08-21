package crypto

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBreakerStopsWhenContextEnds(t *testing.T) {
	b := NewBreaker(2)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := b.Call(ctx, func(callCtx context.Context) error { calls++; cancel(); <-callCtx.Done(); return callCtx.Err() })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("retry continued after cancellation: %d calls", calls)
	}
	if time.Since(time.Now()) < 0 {
		t.Fatal("unreachable")
	}
}
