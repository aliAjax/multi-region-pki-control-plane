package revocation

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	cryptop "example.com/pki-control-plane/internal/crypto"
	"example.com/pki-control-plane/internal/pki/domain"
	"example.com/pki-control-plane/internal/repository"
	"math/big"
	"time"
)

type CRLGenerator struct {
	store repository.Store
	hsm   *cryptop.DevHSM
}

func NewCRLGenerator(s repository.Store, h *cryptop.DevHSM) *CRLGenerator {
	return &CRLGenerator{store: s, hsm: h}
}
func (g *CRLGenerator) Generate(ctx context.Context, caID domain.ID, number int64) (string, error) {
	ca, err := g.store.GetCA(ctx, caID)
	if err != nil {
		return "", err
	}
	issuer, err := cryptop.ParseCertificate(ca.CertificatePEM)
	if err != nil {
		return "", err
	}
	entries := []x509.RevocationListEntry{}
	for _, c := range g.store.ListCertificates(ctx) {
		if c.CAID == caID && c.Status == domain.CertRevoked && c.RevokedAt != nil {
			n := new(big.Int)
			n.SetString(c.Serial, 16)
			entries = append(entries, x509.RevocationListEntry{SerialNumber: n, RevocationTime: *c.RevokedAt, Extensions: []pkix.Extension{}})
		}
	}
	signer, err := g.hsm.Signer(ctx, ca.KeyReference)
	if err != nil {
		return "", err
	}
	tpl := &x509.RevocationList{Number: big.NewInt(number), ThisUpdate: time.Now().UTC(), NextUpdate: time.Now().UTC().Add(24 * time.Hour), RevokedCertificateEntries: entries}
	der, err := x509.CreateRevocationList(rand.Reader, tpl, issuer, signer)
	if err != nil {
		return "", err
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "X509 CRL", Bytes: der})), nil
}
