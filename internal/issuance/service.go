package issuance

import (
	"context"
	stdcrypto "crypto"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"errors"
	cryptop "example.com/pki-control-plane/internal/crypto"
	"example.com/pki-control-plane/internal/pki/domain"
	"example.com/pki-control-plane/internal/repository"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type Service struct {
	store  repository.Store
	signer *cryptop.DevHSM
}

func NewService(s repository.Store, h *cryptop.DevHSM) *Service { return &Service{store: s, signer: h} }

type Request struct {
	Tenant         domain.Tenant
	ProfileID      domain.ID
	CSRPEM         string
	IdempotencyKey string
	Actor          string
}

func (s *Service) Request(ctx context.Context, r Request) (domain.Certificate, error) {
	if c, ok := s.store.FindCertificateByIdempotency(ctx, r.IdempotencyKey, r.Tenant); ok {
		return c, nil
	}
	if r.IdempotencyKey == "" || r.CSRPEM == "" {
		return domain.Certificate{}, errors.New("csr and idempotency key required")
	}
	p, err := s.store.GetProfile(ctx, r.ProfileID)
	if err != nil {
		return domain.Certificate{}, err
	}
	csr, err := cryptop.ParseCSR(r.CSRPEM)
	if err != nil {
		return domain.Certificate{}, fmt.Errorf("parse csr: %w", err)
	}
	if !p.PermitsAlgorithm(domain.ECDSAP256) {
		return domain.Certificate{}, errors.New("csr key algorithm not permitted")
	}
	if err := p.ValidateSANs(csr.DNSNames, csr.IPAddresses, csr.URIs); err != nil {
		return domain.Certificate{}, err
	}
	now := time.Now().UTC()
	c := domain.Certificate{ID: repository.NewID("cert"), Tenant: r.Tenant, CAID: p.CAID, ProfileID: p.ID, CSRPEM: r.CSRPEM, DNSNames: append([]string(nil), csr.DNSNames...), Status: domain.CertRequested, Validity: domain.ValidityWindow{NotBefore: now.Add(-time.Minute), NotAfter: now.Add(p.MaxValidity)}, IdempotencyKey: r.IdempotencyKey, CreatedBy: r.Actor, UpdatedBy: r.Actor, CreatedAt: now, UpdatedAt: now}
	if p.RequireApproval {
		_ = c.Transition(domain.CertPendingApproval, r.Actor, now)
	} else {
		_ = c.Transition(domain.CertApproved, r.Actor, now)
	}
	if err := c.ValidateNew(); err != nil {
		return c, err
	}
	if err := s.store.SaveCertificate(ctx, &c); err != nil {
		return domain.Certificate{}, err
	}
	return c, nil
}
func (s *Service) Issue(ctx context.Context, id, actor string) (domain.Certificate, error) {
	c, err := s.store.GetCertificate(ctx, domain.ID(id))
	if err != nil {
		return c, err
	}
	ca, err := s.store.GetCA(ctx, c.CAID)
	if err != nil {
		return c, err
	}
	if c.Status == domain.CertPendingApproval {
		return c, errors.New("approval required")
	}
	if c.Status != domain.CertApproved && c.Status != domain.CertRenewing {
		return c, fmt.Errorf("cannot issue from %s", c.Status)
	}
	_, pub, err := s.signer.Generate(ctx, string(ca.Algorithm), string(ca.ID))
	if err != nil {
		return c, err
	}
	if err := c.Transition(domain.CertIssuing, actor, time.Now().UTC()); err != nil {
		return c, err
	}
	serial := big.NewInt(time.Now().UnixNano())
	tpl := &x509.Certificate{SerialNumber: serial, Subject: pkix.Name{CommonName: strings.Join(c.DNSNames, ",")}, DNSNames: c.DNSNames, NotBefore: c.Validity.NotBefore, NotAfter: c.Validity.NotAfter, KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}}
	parent, err := cryptop.ParseCertificate(ca.CertificatePEM)
	if err != nil {
		return c, errors.New("ca certificate unavailable")
	}
	der, err := s.signer.Sign(ctx, tpl, parent, pub, ca.KeyReference)
	if err != nil {
		return c, err
	}
	c.Serial = serial.Text(16)
	c.PEM = cryptop.EncodeCertificate(der)
	c.Fingerprint = domain.NewFingerprint(der)
	_ = c.Transition(domain.CertIssued, actor, time.Now().UTC())
	_ = c.Transition(domain.CertPublished, actor, time.Now().UTC())
	if err := s.store.SaveCertificate(ctx, &c); err != nil {
		return c, err
	}
	return c, nil
}
func PublicKeyFingerprint(pub stdcrypto.PublicKey) string {
	b, _ := x509.MarshalPKIXPublicKey(pub)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
