package repository

import (
	"context"
	"errors"
	"example.com/pki-control-plane/internal/pki/domain"
	"fmt"
	"sync"
	"time"
)

var ErrNotFound = errors.New("not found")

type Store interface {
	SaveCA(context.Context, *domain.CertificateAuthority) error
	GetCA(context.Context, domain.ID) (domain.CertificateAuthority, error)
	ListCAs(context.Context) []domain.CertificateAuthority
	SaveProfile(context.Context, *domain.CertificateProfile) error
	GetProfile(context.Context, domain.ID) (domain.CertificateProfile, error)
	SaveCertificate(context.Context, *domain.Certificate) error
	GetCertificate(context.Context, domain.ID) (domain.Certificate, error)
	FindCertificateByIdempotency(context.Context, string, domain.Tenant) (domain.Certificate, bool)
	ListCertificates(context.Context) []domain.Certificate
	SaveAudit(context.Context, domain.AuditRecord) error
	Audits(context.Context) []domain.AuditRecord
	SaveOutbox(context.Context, domain.OutboxEvent) error
	Outbox(context.Context) []domain.OutboxEvent
}
type MemoryStore struct {
	mu       sync.RWMutex
	cas      map[domain.ID]domain.CertificateAuthority
	profiles map[domain.ID]domain.CertificateProfile
	certs    map[domain.ID]domain.Certificate
	audits   []domain.AuditRecord
	outbox   []domain.OutboxEvent
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{cas: map[domain.ID]domain.CertificateAuthority{}, profiles: map[domain.ID]domain.CertificateProfile{}, certs: map[domain.ID]domain.Certificate{}}
}
func (s *MemoryStore) SaveCA(_ context.Context, c *domain.CertificateAuthority) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.cas[c.ID]; ok {
		return errors.New("ca already exists")
	}
	s.cas[c.ID] = *c
	return nil
}
func (s *MemoryStore) GetCA(_ context.Context, id domain.ID) (domain.CertificateAuthority, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.cas[id]
	if !ok {
		return c, fmt.Errorf("ca %s: %w", id, ErrNotFound)
	}
	return c, nil
}
func (s *MemoryStore) ListCAs(_ context.Context) []domain.CertificateAuthority {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r := make([]domain.CertificateAuthority, 0, len(s.cas))
	for _, c := range s.cas {
		r = append(r, c)
	}
	return r
}
func (s *MemoryStore) SaveProfile(_ context.Context, p *domain.CertificateProfile) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.profiles[p.ID]; ok {
		return errors.New("profile already exists")
	}
	s.profiles[p.ID] = *p
	return nil
}
func (s *MemoryStore) GetProfile(_ context.Context, id domain.ID) (domain.CertificateProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.profiles[id]
	if !ok {
		return p, fmt.Errorf("profile %s: %w", id, ErrNotFound)
	}
	return p, nil
}
func (s *MemoryStore) SaveCertificate(_ context.Context, c *domain.Certificate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.certs {
		if v.Fingerprint != "" && v.Fingerprint == c.Fingerprint {
			return errors.New("fingerprint already exists")
		}
		if v.IdempotencyKey == c.IdempotencyKey && v.Tenant.Key() == c.Tenant.Key() {
			return errors.New("idempotency key already exists")
		}
	}
	s.certs[c.ID] = c.Clone()
	return nil
}
func (s *MemoryStore) GetCertificate(_ context.Context, id domain.ID) (domain.Certificate, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.certs[id]
	if !ok {
		return c, fmt.Errorf("certificate %s: %w", id, ErrNotFound)
	}
	return c.Clone(), nil
}
func (s *MemoryStore) FindCertificateByIdempotency(_ context.Context, key string, t domain.Tenant) (domain.Certificate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, c := range s.certs {
		if c.IdempotencyKey == key && c.Tenant.Key() == t.Key() {
			return c.Clone(), true
		}
	}
	return domain.Certificate{}, false
}
func (s *MemoryStore) ListCertificates(_ context.Context) []domain.Certificate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r := make([]domain.Certificate, 0, len(s.certs))
	for _, c := range s.certs {
		r = append(r, c.Clone())
	}
	return r
}
func (s *MemoryStore) SaveAudit(_ context.Context, a domain.AuditRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audits = append(s.audits, a)
	return nil
}
func (s *MemoryStore) Audits(_ context.Context) []domain.AuditRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.AuditRecord(nil), s.audits...)
}
func (s *MemoryStore) SaveOutbox(_ context.Context, e domain.OutboxEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.outbox = append(s.outbox, e)
	return nil
}
func (s *MemoryStore) Outbox(_ context.Context) []domain.OutboxEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]domain.OutboxEvent(nil), s.outbox...)
}
func NewID(prefix string) domain.ID {
	return domain.ID(prefix + "-" + time.Now().UTC().Format("20060102150405.000000000"))
}
