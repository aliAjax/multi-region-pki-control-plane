package notification

import (
	"context"
	"errors"
	"time"
)

type MockProvider struct {
	Fail  bool
	Delay time.Duration
}

func (p MockProvider) Send(ctx context.Context, m Message) (string, error) {
	if p.Delay > 0 {
		t := time.NewTimer(p.Delay)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-t.C:
		}
	}
	if p.Fail {
		return "", errors.New("mock provider failure")
	}
	return "mock-" + m.ID, nil
}
