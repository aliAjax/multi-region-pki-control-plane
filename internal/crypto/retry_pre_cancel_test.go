package crypto

import (
	"context"
	"testing"
)

func TestBreakerRejectsEndedContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := NewBreaker(2).Call(ctx, func(context.Context) error { calls++; return nil })
	if err != context.Canceled || calls != 0 {
		t.Fatalf("pre-cancel ignored: %d %v", calls, err)
	}
}
