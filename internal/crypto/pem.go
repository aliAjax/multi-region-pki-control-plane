package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

var ErrInvalidCertificatePEM = errors.New("invalid certificate PEM")

func EncodeCertificate(der []byte) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}
func ParseCertificate(s string) (*x509.Certificate, error) {
	b, _ := pem.Decode([]byte(s))
	if b == nil || b.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("%w", ErrInvalidCertificatePEM)
	}
	c, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCertificatePEM, err)
	}
	return c, nil
}
func ParseCSR(s string) (*x509.CertificateRequest, error) {
	b, _ := pem.Decode([]byte(s))
	if b == nil || b.Type != "CERTIFICATE REQUEST" {
		return nil, errors.New("invalid CSR PEM")
	}
	csr, e := x509.ParseCertificateRequest(b.Bytes)
	if e != nil {
		return nil, e
	}
	if e = csr.CheckSignature(); e != nil {
		return nil, e
	}
	return csr, nil
}
func EncodeCSR(der []byte) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der}))
}
func VerifyChain(leaf *x509.Certificate, roots, inters *x509.CertPool, dns string) ([][]*x509.Certificate, error) {
	return leaf.Verify(x509.VerifyOptions{Roots: roots, Intermediates: inters, DNSName: dns})
}
