package revocation

import (
	"context"
	"errors"
	"example.com/pki-control-plane/internal/pki/domain"
	"example.com/pki-control-plane/internal/repository"
	"testing"
)

var errAuditUnavailable = errors.New("audit unavailable")

type auditFailingStore struct{ repository.Store }

func (auditFailingStore) SaveAudit(context.Context, domain.AuditRecord) error {
	return errAuditUnavailable
}
func TestRevokePreservesAuditError(t *testing.T) {
	s := repository.NewMemoryStore()
	c := &domain.Certificate{ID: "cert-audit", Tenant: domain.Tenant{Organization: "o", Service: "s", Environment: "p", Region: "r"}, Status: domain.CertPublished}
	_ = s.SaveCertificate(context.Background(), c)
	_, e := NewService(auditFailingStore{Store: s}).Revoke(context.Background(), string(c.ID), ReasonKeyCompromise, "a", 0)
	if !errors.Is(e, errAuditUnavailable) {
		t.Fatalf("audit error lost: %v", e)
	}
}
