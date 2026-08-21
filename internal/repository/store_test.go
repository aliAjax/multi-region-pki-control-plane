package repository

import (
	"context"
	"sync"
	"testing"

	"example.com/pki-control-plane/internal/pki/domain"
)

func TestMemoryStoreConcurrentCertificateSnapshot(t *testing.T) {
	s := NewMemoryStore()
	dns := []string{"edge.example.test"}
	c := &domain.Certificate{ID: "cert-1", CAID: "ca-1", ProfileID: "profile-1", Tenant: domain.Tenant{Organization: "acme", Service: "edge", Environment: "prod", Region: "jp"}, CSRPEM: "csr", IdempotencyKey: "idem-1", DNSNames: dns}
	if err := s.SaveCertificate(context.Background(), c); err != nil { t.Fatal(err) }
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); for i := 0; i < 200; i++ { dns[0] = "edge.example.test" } }()
	go func() { defer wg.Done(); for i := 0; i < 200; i++ { got := s.ListCertificates(context.Background()); if len(got) != 1 || len(got[0].DNSNames) != 1 { t.Fatalf("incomplete snapshot: %#v", got) } } }()
	wg.Wait()
}
