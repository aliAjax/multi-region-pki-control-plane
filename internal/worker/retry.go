package worker

import (
	"context"
	"math"
	"time"
)

type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

func (p RetryPolicy) Delay(attempt int) time.Duration {
	if p.BaseDelay <= 0 {
		p.BaseDelay = time.Second
	}
	if p.MaxDelay <= 0 {
		p.MaxDelay = time.Minute
	}
	d := float64(p.BaseDelay) * math.Pow(2, float64(attempt))
	if d > float64(p.MaxDelay) {
		d = float64(p.MaxDelay)
	}
	return time.Duration(d)
}
func (p RetryPolicy) Run(ctx context.Context, fn func(context.Context) error) error {
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 3
	}
	var err error
	for i := 0; i < p.MaxAttempts; i++ {
		if err = fn(ctx); err == nil {
			return nil
		}
		time.Sleep(p.Delay(i))
	}
	return err
}
