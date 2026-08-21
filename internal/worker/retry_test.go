package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryPolicyStopsWhenContextEnds(t *testing.T) {
	c, cancel := context.WithCancel(context.Background())
	n := 0
	e := RetryPolicy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}.Run(c, func(context.Context) error { n++; cancel(); return errors.New("x") })
	if !errors.Is(e, context.Canceled) || n != 1 {
		t.Fatalf("%d %v", n, e)
	}
}
