package notification

import (
	"context"
	"errors"
	"testing"
	"time"
)

type flakyProvider struct{ calls int }

func (p *flakyProvider) Send(context.Context, Message) (string, error) {
	p.calls++
	if p.calls == 1 {
		return "", errors.New("temporary")
	}
	return "delivered", nil
}

type delayedProvider struct {
	calls     int
	delay     time.Duration
	failFirst bool
}

func (p *delayedProvider) Send(context.Context, Message) (string, error) {
	p.calls++
	time.Sleep(p.delay)
	if p.failFirst && p.calls == 1 {
		return "", errors.New("temporary")
	}
	return "delivered", nil
}

func TestDispatcherRetriesAfterProviderIssue(t *testing.T) {
	p := &flakyProvider{}
	d := NewDispatcher(0)
	if err := d.Register(ChannelEmail, p); err != nil {
		t.Fatal(err)
	}
	m := Message{ID: "m-1", Tenant: "tenant", Channel: ChannelEmail, IdempotencyKey: "same"}
	if _, err := d.Submit(context.Background(), m); err == nil {
		t.Fatal("first delivery should fail")
	}
	got, err := d.Submit(context.Background(), m)
	if err != nil || got != "delivered" {
		t.Fatalf("retry did not reach provider: %q %v", got, err)
	}
}

func TestDispatcherSuppressesDuplicateAfterSuccess(t *testing.T) {
	p := &delayedProvider{delay: 20 * time.Millisecond, failFirst: true}
	d := NewDispatcher(time.Hour)
	if err := d.Register(ChannelEmail, p); err != nil {
		t.Fatal(err)
	}
	m := Message{ID: "m-success", Tenant: "tenant", Channel: ChannelEmail, IdempotencyKey: "success"}
	if _, err := d.Submit(context.Background(), m); err == nil {
		t.Fatal("first delivery should fail")
	}
	if _, err := d.Submit(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Submit(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if p.calls != 2 {
		t.Fatalf("successful message sent %d times", p.calls)
	}
}
