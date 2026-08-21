package notification

import (
	"context"
	"errors"
	"testing"
)

type flakyProvider struct{ calls int }
func (p *flakyProvider) Send(context.Context, Message) (string, error) { p.calls++; if p.calls == 1 { return "", errors.New("temporary") }; return "delivered", nil }

func TestDispatcherRetryAfterProviderFailure(t *testing.T) {
	p := &flakyProvider{}
	d := NewDispatcher(0)
	if err := d.Register(ChannelEmail, p); err != nil { t.Fatal(err) }
	m := Message{ID: "m-1", Tenant: "tenant", Channel: ChannelEmail, IdempotencyKey: "same"}
	if _, err := d.Submit(context.Background(), m); err == nil { t.Fatal("first delivery should fail") }
	got, err := d.Submit(context.Background(), m)
	if err != nil || got != "delivered" { t.Fatalf("retry did not reach provider: %q %v", got, err) }
}
