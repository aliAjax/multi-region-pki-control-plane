package issuance

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	cryptop "example.com/pki-control-plane/internal/crypto"
	"example.com/pki-control-plane/internal/pki/domain"
	"example.com/pki-control-plane/internal/repository"
	"math/big"
	"time"
)

type CAService struct {
	store repository.Store
	hsm   *cryptop.DevHSM
}

func NewCAService(s repository.Store, h *cryptop.DevHSM) *CAService {
	return &CAService{store: s, hsm: h}
}

type CreateCARequest struct {
	Name      string
	Type      domain.CAType
	ParentID  domain.ID
	Tenant    domain.Tenant
	Region    string
	Algorithm domain.KeyAlgorithm
	Validity  time.Duration
	Actor     string
}

func (s *CAService) Create(ctx context.Context, r CreateCARequest) (domain.CertificateAuthority, error) {
	if r.Validity == 0 {
		r.Validity = 10 * 365 * 24 * time.Hour
	}
	if r.Algorithm == "" {
		r.Algorithm = domain.ECDSAP256
	}
	now := time.Now().UTC()
	id := repository.NewID("ca")
	ref, pub, err := s.hsm.Generate(ctx, string(r.Algorithm), string(id))
	if err != nil {
		return domain.CertificateAuthority{}, err
	}
	ca := domain.CertificateAuthority{ID: id, Name: r.Name, Type: r.Type, ParentID: r.ParentID, Tenant: r.Tenant, Region: r.Region, Algorithm: r.Algorithm, KeyReference: ref, Status: domain.CAPending, NotBefore: now.Add(-time.Minute), NotAfter: now.Add(r.Validity), Version: 1, CreatedAt: now, UpdatedAt: now}
	if err := ca.Validate(); err != nil {
		return ca, err
	}
	tpl := &x509.Certificate{SerialNumber: big.NewInt(now.UnixNano()), Subject: pkix.Name{CommonName: r.Name}, NotBefore: ca.NotBefore, NotAfter: ca.NotAfter, IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature, MaxPathLen: 1}
	parent := tpl
	keyRef := ref
	if r.Type == domain.IntermediateCA {
		p, err := s.store.GetCA(ctx, r.ParentID)
		if err != nil {
			return ca, err
		}
		parent, err = cryptop.ParseCertificate(p.CertificatePEM)
		if err != nil {
			return ca, err
		}
		keyRef = p.KeyReference
	}
	der, err := s.hsm.Sign(ctx, tpl, parent, pub, keyRef)
	if err != nil {
		return ca, err
	}
	ca.CertificatePEM = cryptop.EncodeCertificate(der)
	if err := ca.Transition(domain.CAActive, now); err != nil {
		return ca, err
	}
	if err := s.store.SaveCA(ctx, &ca); err != nil {
		return ca, err
	}
	_ = s.store.SaveAudit(ctx, domain.AuditRecord{ID: repository.NewID("audit"), AggregateType: "ca", AggregateID: ca.ID, Action: "created", To: string(ca.Status), Actor: r.Actor, TenantKey: r.Tenant.Key(), CreatedAt: now})
	return ca, nil
}
func (s *CAService) Rotate(ctx context.Context, id domain.ID, actor string) (domain.CertificateAuthority, error) {
	ca, err := s.store.GetCA(ctx, id)
	if err != nil {
		return ca, err
	}
	if ca.Status != domain.CAActive {
		return ca, errors.New("only active CA can rotate")
	}
	if err := ca.Transition(domain.CARotating, time.Now().UTC()); err != nil {
		return ca, err
	}
	return ca, nil
}
