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
		time.Sleep(p.Delay)
	}
	if p.Fail {
		return "", errors.New("mock provider failure")
	}
	return "mock-" + m.ID, nil
}
