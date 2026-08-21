package challenge

import (
	"errors"
	"sync"
	"time"
)

type Lease struct {
	Key       string
	Owner     string
	Token     uint64
	ExpiresAt time.Time
}
type LeaseManager struct {
	mu     sync.Mutex
	leases map[string]Lease
	tokens map[string]uint64
}

func NewLeaseManager() *LeaseManager {
	return &LeaseManager{leases: map[string]Lease{}, tokens: map[string]uint64{}}
}
func (m *LeaseManager) Acquire(key, owner string, ttl time.Duration, now time.Time) (Lease, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if l, ok := m.leases[key]; ok && now.Before(l.ExpiresAt) && l.Owner != owner {
		return Lease{}, errors.New("challenge lease held")
	}
	m.tokens[key]++
	l := Lease{Key: key, Owner: owner, Token: m.tokens[key], ExpiresAt: now.Add(ttl)}
	m.leases[key] = l
	return l, nil
}
func (m *LeaseManager) Renew(l Lease, ttl time.Duration, now time.Time) (Lease, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.leases[l.Key]
	if !ok || cur.Owner != l.Owner || cur.Token != l.Token {
		return Lease{}, errors.New("stale fencing token")
	}
	if !now.Before(cur.ExpiresAt) {
		return Lease{}, errors.New("lease expired")
	}
	cur.ExpiresAt = now.Add(ttl)
	m.leases[l.Key] = cur
	return cur, nil
}
func (m *LeaseManager) Release(l Lease) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cur, ok := m.leases[l.Key]
	if !ok || cur.Owner != l.Owner || cur.Token != l.Token {
		return errors.New("stale fencing token")
	}
	delete(m.leases, l.Key)
	return nil
}
