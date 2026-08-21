package revocation

import (
	"context"
	"errors"
	"example.com/pki-control-plane/internal/pki/domain"
	"example.com/pki-control-plane/internal/repository"
	"testing"
)

var errCAUnavailable = errors.New("ca repository unavailable")

type crlFailingStore struct{ repository.Store }

func (crlFailingStore) GetCA(context.Context, domain.ID) (domain.CertificateAuthority, error) {
	return domain.CertificateAuthority{}, errCAUnavailable
}
func TestCRLGeneratePreservesRepositoryError(t *testing.T) {
	g := NewCRLGenerator(crlFailingStore{Store: repository.NewMemoryStore()}, nil)
	_, e := g.Generate(context.Background(), "ca", 1)
	if !errors.Is(e, errCAUnavailable) {
		t.Fatalf("lost: %v", e)
	}
}

func TestCRLGenerateClassifiesIssuerParseError(t *testing.T) {
	store := repository.NewMemoryStore()
	_ = store.SaveCA(context.Background(), &domain.CertificateAuthority{ID: "ca-bad", Name: "ca", Type: domain.RootCA, Tenant: domain.Tenant{Organization: "o", Service: "s", Environment: "p", Region: "r"}, CertificatePEM: "bad"})
	g := NewCRLGenerator(store, nil)
	_, e := g.Generate(context.Background(), "ca-bad", 1)
	if !errors.Is(e, ErrInvalidIssuerCertificate) {
		t.Fatalf("issuer error lost: %v", e)
	}
}
