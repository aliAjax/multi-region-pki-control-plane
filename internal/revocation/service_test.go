package revocation

import (
	"context"
	"errors"
	"testing"

	"example.com/pki-control-plane/internal/pki/domain"
	"example.com/pki-control-plane/internal/repository"
)

type failingStore struct{ repository.Store }

func (f failingStore) SaveOutbox(context.Context, domain.OutboxEvent) error {
	return errors.New("outbox unavailable")
}

func TestRevokePropagatesAuditAndOutboxFailure(t *testing.T) {
	base := repository.NewMemoryStore()
	c := &domain.Certificate{ID: "cert-1", CAID: "ca-1", ProfileID: "profile-1", Tenant: domain.Tenant{Organization: "acme", Service: "edge", Environment: "prod", Region: "jp"}, CSRPEM: "csr", IdempotencyKey: "idem-1", Status: domain.CertPublished}
	if err := base.SaveCertificate(context.Background(), c); err != nil {
		t.Fatal(err)
	}
	s := NewService(failingStore{Store: base})
	_, err := s.Revoke(context.Background(), "cert-1", ReasonKeyCompromise, "operator", 0)
	if err == nil || !errors.Is(err, ErrOutboxUnavailable) {
		t.Fatalf("expected outbox failure, got %v", err)
	}
}
