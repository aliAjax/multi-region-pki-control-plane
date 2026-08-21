package crypto

import (
	"context"
	stdcrypto "crypto"
	"crypto/x509"
	"reflect"
)

func signerIsNil(s stdcrypto.Signer) bool {
	if s == nil {
		return true
	}
	v := reflect.ValueOf(s)
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
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
