package domain

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

type nonceEntry struct {
	expires time.Time
	used    bool
}
type NonceStore struct {
	mu      sync.Mutex
	entries map[string]nonceEntry
	ttl     time.Duration
	max     int
}

func NewNonceStore(ttl time.Duration, max int) *NonceStore {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	if max < 100 {
		max = 100
	}
	return &NonceStore{entries: map[string]nonceEntry{}, ttl: ttl, max: max}
}
func (n *NonceStore) Issue(now time.Time) (string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.cleanup(now)
	if len(n.entries) >= n.max {
		return "", errors.New("nonce capacity exhausted")
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	v := base64.RawURLEncoding.EncodeToString(b)
	n.entries[v] = nonceEntry{expires: now.Add(n.ttl)}
	return v, nil
}
func (n *NonceStore) Consume(value string, now time.Time) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	e, ok := n.entries[value]
	if !ok {
		return errors.New("unknown nonce")
	}
	if e.used {
		return errors.New("nonce replayed")
	}
	if !now.Before(e.expires) {
		delete(n.entries, value)
		return errors.New("nonce expired")
	}
	e.used = true
	n.entries[value] = e
	return nil
}
func (n *NonceStore) cleanup(now time.Time) {
	for k, e := range n.entries {
		if e.used || !now.Before(e.expires) {
			delete(n.entries, k)
		}
	}
}
