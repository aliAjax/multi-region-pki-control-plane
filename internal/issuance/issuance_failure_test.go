package issuance

import (
	"context"
	"errors"
	"example.com/pki-control-plane/internal/crypto"
	"example.com/pki-control-plane/internal/pki/domain"
	"example.com/pki-control-plane/internal/repository"
	"testing"
)

func TestIssuanceFailureRollsBackState(t *testing.T) {
	s := repository.NewMemoryStore()
	h := crypto.NewSignerFactory(true)
	r, _, _ := h.Generate(context.Background(), "ECDSA-P256", "rootkey")
	ten := domain.Tenant{Organization: "o", Service: "s", Environment: "p", Region: "r"}
	_ = s.SaveCA(context.Background(), &domain.CertificateAuthority{ID: "ca", Name: "ca", Type: domain.RootCA, Tenant: ten, Algorithm: domain.ECDSAP256, KeyReference: r, Status: domain.CAActive, CertificatePEM: "bad"})
	_ = s.SaveCertificate(context.Background(), &domain.Certificate{ID: "c", CAID: "ca", ProfileID: "p", Tenant: ten, CSRPEM: "x", IdempotencyKey: "i", Status: domain.CertApproved})
	c, e := NewService(s, h).Issue(context.Background(), "c", "w")
	if e == nil || c.Status != domain.CertApproved {
		t.Fatalf("state %s err %v", c.Status, e)
	}
}

func TestCACreatePreservesParentCertificateError(t *testing.T) {
	store := repository.NewMemoryStore()
	tenant := domain.Tenant{Organization: "o", Service: "s", Environment: "p", Region: "r"}
	_ = store.SaveCA(context.Background(), &domain.CertificateAuthority{ID: "parent", Name: "parent", Type: domain.RootCA, Tenant: tenant, CertificatePEM: "bad"})
	hsm := crypto.NewSignerFactory(true)
	_, err := NewCAService(store, hsm).Create(context.Background(), CreateCARequest{Name: "child", Type: domain.IntermediateCA, ParentID: "parent", Tenant: tenant})
	if !errors.Is(err, ErrParentCertificateUnavailable) {
		t.Fatalf("parent error lost: %v", err)
	}
}

func TestCACreateRespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	store := repository.NewMemoryStore()
	hsm := crypto.NewSignerFactory(false)
	_, err := NewCAService(store, hsm).Create(ctx, CreateCARequest{Name: "cancelled", Type: domain.RootCA, Tenant: domain.Tenant{Organization: "o", Service: "s", Environment: "p", Region: "r"}})
	if err != context.Canceled {
		t.Fatalf("cancellation ignored: %v", err)
	}
}
