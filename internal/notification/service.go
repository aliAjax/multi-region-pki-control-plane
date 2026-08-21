package notification

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Channel string

const (
	ChannelEmail   Channel = "email"
	ChannelWebhook Channel = "webhook"
	ChannelSMS     Channel = "sms"
)

type Message struct {
	ID             string
	Tenant         string
	Channel        Channel
	Destination    string
	TemplateID     string
	Payload        map[string]string
	IdempotencyKey string
	Attempt        int
	CreatedAt      time.Time
}
type Provider interface {
	Send(context.Context, Message) (string, error)
}
type ProviderResult struct {
	Accepted   bool
	ExternalID string
	Retryable  bool
	Code       string
}
type Dispatcher struct {
	mu        sync.Mutex
	providers map[Channel]Provider
	seen      map[string]time.Time
	dedupe    time.Duration
}

func NewDispatcher(dedupe time.Duration) *Dispatcher {
	if dedupe <= 0 {
		dedupe = 24 * time.Hour
	}
	return &Dispatcher{providers: map[Channel]Provider{}, seen: map[string]time.Time{}, dedupe: dedupe}
}
func (d *Dispatcher) Register(c Channel, p Provider) error {
	if p == nil {
		return errors.New("nil provider")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.providers[c] = p
	return nil
}
func (d *Dispatcher) Submit(ctx context.Context, m Message) (string, error) {
	if m.IdempotencyKey == "" {
		return "", errors.New("idempotency key required")
	}
	d.mu.Lock()
	if t, ok := d.seen[m.Tenant+":"+m.IdempotencyKey]; ok && time.Since(t) < d.dedupe {
		d.mu.Unlock()
		return m.ID, nil
	}
	p := d.providers[m.Channel]
	d.mu.Unlock()
	if p == nil {
		return "", fmt.Errorf("no provider for %s", m.Channel)
	}
	external, err := p.Send(ctx, m)
	if err == nil {
		d.mu.Lock()
		d.seen[m.Tenant+":"+m.IdempotencyKey] = time.Now()
		d.mu.Unlock()
	}
	return external, err
}

type Receipt struct {
	Provider   string
	MessageID  string
	EventID    string
	Sequence   int64
	Type       string
	Signature  string
	Body       string
	ReceivedAt time.Time
}
type ReceiptStore struct {
	mu     sync.Mutex
	events map[string]int64
}

func NewReceiptStore() *ReceiptStore { return &ReceiptStore{events: map[string]int64{}} }
func (s *ReceiptStore) Accept(r Receipt, secret string) error {
	if !VerifySignature(r.Body, r.Signature, secret) {
		return errors.New("invalid webhook signature")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := r.Provider + ":" + r.MessageID
	if old, ok := s.events[key]; ok && r.Sequence <= old {
		return nil
	}
	s.events[key] = r.Sequence
	return nil
}
func VerifySignature(body, sig, secret string) bool {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(body))
	expected := hex.EncodeToString(m.Sum(nil))
	return strings.EqualFold(expected, sig)
}
