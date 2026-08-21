package crypto

import (
	"context"
	stdcrypto "crypto"
	"crypto/x509"
)

func signerIsNil(s stdcrypto.Signer) bool {
	if s == nil {
		return true
	}
	return false
}

type Signer interface {
	Signer(ctx context.Context, keyRef string) (stdcrypto.Signer, error)
	Available(ctx context.Context, keyRef string) error
	Destroy()
}
type CertificateSigner interface {
	Sign(ctx context.Context, template, parent *x509.Certificate, publicKey any, keyRef string) ([]byte, error)
}
type Capabilities struct {
	Algorithms     []string
	HardwareBacked bool
	Attestation    bool
}
type KeyProvider interface {
	Generate(ctx context.Context, algorithm string, label string) (string, stdcrypto.PublicKey, error)
	Delete(ctx context.Context, keyRef string) error
	Capabilities(ctx context.Context) (Capabilities, error)
}
