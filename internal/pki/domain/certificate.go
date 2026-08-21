package domain

import (
	"errors"
	"fmt"
	"time"
)

type CertificateStatus string

const (
	CertRequested       CertificateStatus = "requested"
	CertPendingApproval CertificateStatus = "pending_approval"
	CertApproved        CertificateStatus = "approved"
	CertIssuing         CertificateStatus = "issuing"
	CertIssued          CertificateStatus = "issued"
	CertPublished       CertificateStatus = "published"
	CertRenewing        CertificateStatus = "renewing"
	CertRevoked         CertificateStatus = "revoked"
	CertExpired         CertificateStatus = "expired"
	CertCancelled       CertificateStatus = "cancelled"
	CertArchived        CertificateStatus = "archived"
	CertFailed          CertificateStatus = "failed"
)

type Certificate struct {
	ID                   ID                `json:"id"`
	Tenant               Tenant            `json:"tenant"`
	CAID                 ID                `json:"ca_id"`
	ProfileID            ID                `json:"profile_id"`
	Serial               string            `json:"serial"`
	Fingerprint          Fingerprint       `json:"fingerprint"`
	PublicKeyFingerprint Fingerprint       `json:"public_key_fingerprint"`
	CSRPEM               string            `json:"csr_pem,omitempty"`
	PEM                  string            `json:"pem,omitempty"`
	ChainPEM             string            `json:"chain_pem,omitempty"`
	DNSNames             []string          `json:"dns_names"`
	Status               CertificateStatus `json:"status"`
	Validity             ValidityWindow    `json:"validity"`
	IdempotencyKey       string            `json:"idempotency_key"`
	RenewalOf            ID                `json:"renewal_of,omitempty"`
	RevocationReason     string            `json:"revocation_reason,omitempty"`
	RevokedAt            *time.Time        `json:"revoked_at,omitempty"`
	Version              int64             `json:"version"`
	CreatedBy            string            `json:"created_by"`
	UpdatedBy            string            `json:"updated_by"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

// Clone returns an ownership-safe snapshot for repository boundaries.
func (c Certificate) Clone() Certificate {
	c.DNSNames = append([]string(nil), c.DNSNames...)
	return c
}

func (c Certificate) ValidateNew() error {
	if c.ID == "" || c.CAID == "" || c.ProfileID == "" {
		return errors.New("certificate id, ca and profile required")
	}
	if c.IdempotencyKey == "" {
		return errors.New("idempotency key required")
	}
	if c.CSRPEM == "" {
		return errors.New("csr required")
	}
	return c.Tenant.Validate()
}
func (c *Certificate) Transition(next CertificateStatus, actor string, now time.Time) error {
	transitions := map[CertificateStatus]map[CertificateStatus]bool{CertRequested: {CertPendingApproval: true, CertApproved: true, CertCancelled: true, CertFailed: true}, CertPendingApproval: {CertApproved: true, CertCancelled: true, CertFailed: true}, CertApproved: {CertIssuing: true, CertCancelled: true, CertFailed: true}, CertIssuing: {CertIssued: true, CertFailed: true}, CertIssued: {CertPublished: true, CertRevoked: true, CertExpired: true}, CertPublished: {CertRenewing: true, CertRevoked: true, CertExpired: true}, CertRenewing: {CertPublished: true, CertRevoked: true, CertExpired: true, CertFailed: true}, CertExpired: {CertArchived: true}, CertRevoked: {CertArchived: true}, CertCancelled: {CertArchived: true}, CertFailed: {CertApproved: true, CertArchived: true}}
	if !transitions[c.Status][next] {
		return fmt.Errorf("invalid certificate transition %s to %s", c.Status, next)
	}
	c.Status = next
	c.UpdatedBy = actor
	c.UpdatedAt = now
	c.Version++
	if next == CertRevoked {
		c.RevokedAt = &now
	}
	return nil
}

type AuditRecord struct {
	ID            ID        `json:"id"`
	AggregateType string    `json:"aggregate_type"`
	AggregateID   ID        `json:"aggregate_id"`
	TenantKey     string    `json:"tenant_key"`
	Action        string    `json:"action"`
	From          string    `json:"from"`
	To            string    `json:"to"`
	Actor         string    `json:"actor"`
	Detail        string    `json:"detail"`
	PreviousHash  string    `json:"previous_hash"`
	Hash          string    `json:"hash"`
	CreatedAt     time.Time `json:"created_at"`
}
type OutboxEvent struct {
	ID          ID         `json:"id"`
	AggregateID ID         `json:"aggregate_id"`
	Type        string     `json:"type"`
	Payload     []byte     `json:"payload"`
	Attempts    int        `json:"attempts"`
	AvailableAt time.Time  `json:"available_at"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
