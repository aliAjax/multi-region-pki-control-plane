package notification

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestProviderStopsOnContextEnd(t *testing.T) {
	c, x := context.WithCancel(context.Background())
	x()
	start := time.Now()
	_, e := (MockProvider{Delay: 100 * time.Millisecond}).Send(c, Message{})
	if !errors.Is(e, context.Canceled) || time.Since(start) > 50*time.Millisecond {
		t.Fatalf("cancel: %v", e)
	}
	if _, e = Render("{{serial}} {{tenant}}", map[string]string{"serial": "1"}); e == nil {
		t.Fatal("missing accepted")
	}
}
