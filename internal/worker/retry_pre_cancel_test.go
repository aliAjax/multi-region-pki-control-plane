package worker

import (
	"context"
	"testing"
)

func TestRetryRejectsEndedContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := RetryPolicy{MaxAttempts: 3}.Run(ctx, func(context.Context) error { calls++; return nil })
	if err != context.Canceled || calls != 0 {
		t.Fatalf("pre-cancel ignored: %d %v", calls, err)
	}
}
