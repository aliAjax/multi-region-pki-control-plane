package revocation

import (
	"context"
	"errors"
	"example.com/pki-control-plane/internal/pki/domain"
	"example.com/pki-control-plane/internal/repository"
	"fmt"
	"time"
)

type Reason string

const (
	ReasonUnspecified        Reason = "unspecified"
	ReasonKeyCompromise      Reason = "key_compromise"
	ReasonCACompromise       Reason = "ca_compromise"
	ReasonAffiliationChanged Reason = "affiliation_changed"
	ReasonSuperseded         Reason = "superseded"
	ReasonCessation          Reason = "cessation_of_operation"
	ReasonPrivilegeWithdrawn Reason = "privilege_withdrawn"
)

func (r Reason) Valid() bool {
	switch r {
	case ReasonUnspecified, ReasonKeyCompromise, ReasonCACompromise, ReasonAffiliationChanged, ReasonSuperseded, ReasonCessation, ReasonPrivilegeWithdrawn:
		return true
	}
	return false
}

type Service struct{ store repository.Store }

var ErrOutboxUnavailable = errors.New("outbox unavailable")

func NewService(s repository.Store) *Service { return &Service{store: s} }
func (s *Service) Revoke(ctx context.Context, id string, reason Reason, actor string, expected int64) (domain.Certificate, error) {
	if !reason.Valid() {
		return domain.Certificate{}, errors.New("invalid revocation reason")
	}
	c, err := s.store.GetCertificate(ctx, domain.ID(id))
	if err != nil {
		return c, err
	}
	if expected > 0 && c.Version != expected {
		return c, fmt.Errorf("version mismatch: have %d", c.Version)
	}
	if c.Status == domain.CertRevoked {
		return c, nil
	}
	if c.Status == domain.CertCancelled || c.Status == domain.CertArchived {
		return c, errors.New("certificate cannot be revoked")
	}
	from := c.Status
	now := time.Now().UTC()
	c.RevocationReason = string(reason)
	if err := c.Transition(domain.CertRevoked, actor, now); err != nil {
		return c, err
	}
	if err := s.store.SaveAudit(ctx, domain.AuditRecord{ID: repository.NewID("audit"), AggregateType: "certificate", AggregateID: c.ID, Action: "revoked", From: string(from), To: string(c.Status), Actor: actor, Detail: string(reason), TenantKey: c.Tenant.Key(), CreatedAt: now}); err != nil {
		return c, fmt.Errorf("save revocation audit: %w", err)
	}
	if err := s.store.SaveOutbox(ctx, domain.OutboxEvent{ID: repository.NewID("event"), AggregateID: c.ID, Type: "certificate.revoked", Payload: []byte(fmt.Sprintf(`{"id":%q,"serial":%q}`, c.ID, c.Serial)), AvailableAt: now, CreatedAt: now}); err != nil {
		return c, fmt.Errorf("%w: %v", ErrOutboxUnavailable, err)
	}
	return c, nil
}
func (s *Service) Batch(ctx context.Context, ids []string, reason Reason, actor string) map[string]error {
	result := make(map[string]error, len(ids))
	for _, id := range ids {
		_, err := s.Revoke(ctx, id, reason, actor, 0)
		result[id] = err
	}
	return result
}

type OCSPStatus struct {
	Serial     string     `json:"serial"`
	Status     string     `json:"status"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	Reason     string     `json:"reason,omitempty"`
	ProducedAt time.Time  `json:"produced_at"`
	NextUpdate time.Time  `json:"next_update"`
}

func (s *Service) Check(ctx context.Context, serial string) (OCSPStatus, error) {
	for _, c := range s.store.ListCertificates(ctx) {
		if c.Serial == serial {
			now := time.Now().UTC()
			status := "good"
			if c.Status == domain.CertRevoked {
				status = "revoked"
			} else if c.Status == domain.CertExpired || !c.ActiveAt(now) {
				status = "expired"
			}
			return OCSPStatus{Serial: serial, Status: status, RevokedAt: c.RevokedAt, Reason: c.RevocationReason, ProducedAt: now, NextUpdate: now.Add(time.Hour)}, nil
		}
	}
	return OCSPStatus{Serial: serial, Status: "unknown", ProducedAt: time.Now().UTC()}, nil
}
