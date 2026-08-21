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
	if err := s.SaveCertificate(context.Background(), c); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 200; i++ {
			dns[0] = "edge.example.test"
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 200; i++ {
			got := s.ListCertificates(context.Background())
			if len(got) != 1 || len(got[0].DNSNames) != 1 || got[0].DNSNames[0] == "" {
				t.Fatalf("incomplete snapshot: %#v", got)
			}
			clone := got[0].Clone()
			clone.DNSNames[0] = "mutated.example.test"
			if got[0].DNSNames[0] == "mutated.example.test" {
				t.Fatal("certificate clone exposed the source slice")
			}
		}
	}()
	close(start)
	wg.Wait()
}

func TestMemoryStoreCopiesCertificateInputs(t *testing.T) {
	s := NewMemoryStore()
	dns := []string{"edge.example.test"}
	c := &domain.Certificate{ID: "cert-input", CAID: "ca", ProfileID: "profile", Tenant: domain.Tenant{Organization: "o", Service: "s", Environment: "p", Region: "r"}, CSRPEM: "csr", IdempotencyKey: "idem-input", DNSNames: dns}
	if err := s.SaveCertificate(context.Background(), c); err != nil {
		t.Fatal(err)
	}
	dns[0] = "changed.example.test"
	got, err := s.GetCertificate(context.Background(), c.ID)
	if err != nil || got.DNSNames[0] != "edge.example.test" {
		t.Fatalf("input slice escaped: %#v %v", got, err)
	}
}

func TestMemoryStoreGetCertificateReturnsCopy(t *testing.T) {
	s, c := seededCertificateStore(t, "get")
	first, _ := s.GetCertificate(context.Background(), c.ID)
	first.DNSNames[0] = "changed.example.test"
	second, _ := s.GetCertificate(context.Background(), c.ID)
	if second.DNSNames[0] != "edge.example.test" {
		t.Fatalf("GetCertificate leaked state: %#v", second)
	}
}

func TestMemoryStoreIdempotencyLookupReturnsCopy(t *testing.T) {
	s, c := seededCertificateStore(t, "find")
	first, _ := s.FindCertificateByIdempotency(context.Background(), c.IdempotencyKey, c.Tenant)
	first.DNSNames[0] = "changed.example.test"
	second, _ := s.FindCertificateByIdempotency(context.Background(), c.IdempotencyKey, c.Tenant)
	if second.DNSNames[0] != "edge.example.test" {
		t.Fatalf("idempotency lookup leaked state: %#v", second)
	}
}

func TestMemoryStoreListCertificatesReturnsCopy(t *testing.T) {
	s, _ := seededCertificateStore(t, "list")
	first := s.ListCertificates(context.Background())
	first[0].DNSNames[0] = "changed.example.test"
	second := s.ListCertificates(context.Background())
	if second[0].DNSNames[0] != "edge.example.test" {
		t.Fatalf("ListCertificates leaked state: %#v", second)
	}
}

func TestCloneOwnsDNSNames(t *testing.T) {
	original := domain.Certificate{DNSNames: []string{"edge.example.test"}}
	clone := original.Clone()
	clone.DNSNames[0] = "changed.example.test"
	if original.DNSNames[0] != "edge.example.test" {
		t.Fatalf("Clone leaked source state: %#v", original)
	}
}

func seededCertificateStore(t *testing.T, suffix string) (*MemoryStore, *domain.Certificate) {
	t.Helper()
	s := NewMemoryStore()
	c := &domain.Certificate{ID: domain.ID("cert-" + suffix), CAID: "ca", ProfileID: "profile", Tenant: domain.Tenant{Organization: "o", Service: "s", Environment: "p", Region: "r"}, CSRPEM: "csr", IdempotencyKey: "idem-" + suffix, DNSNames: []string{"edge.example.test"}}
	if err := s.SaveCertificate(context.Background(), c); err != nil {
		t.Fatal(err)
	}
	return s, c
}
