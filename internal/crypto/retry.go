package crypto

import (
	"context"
	"errors"
	"time"
)

type Breaker struct {
	failures  int
	threshold int
	openUntil time.Time
}

func NewBreaker(n int) *Breaker {
	if n < 1 {
		n = 3
	}
	return &Breaker{threshold: n}
}
func (b *Breaker) Call(ctx context.Context, fn func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	now := time.Now()
	if now.Before(b.openUntil) {
		return errors.New("signing circuit open")
	}
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		if err = ctx.Err(); err != nil {
			return err
		}
		callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err = fn(callCtx)
		cancel()
		if err == nil {
			b.failures = 0
			return nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		select {
		case <-time.After(time.Duration(1<<attempt) * 50 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	b.failures++
	if b.failures >= b.threshold {
		b.openUntil = now.Add(30 * time.Second)
	}
	return err
}
