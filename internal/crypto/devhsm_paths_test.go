package crypto

import (
	"context"
	"crypto/ecdsa"
	"testing"
)

func TestDevHSMSignerRejectsTypedNil(t *testing.T) {
	d := NewSignerFactory(true)
	d.keys["nil"] = (*ecdsa.PrivateKey)(nil)
	if s, e := d.Signer(context.Background(), "nil"); e == nil || s != nil || !signerIsNil(s) {
		t.Fatalf("typed nil escaped: %v %v", s, e)
	}
}

func TestDevHSMDeleteRejectsTypedNil(t *testing.T) {
	d := NewSignerFactory(true)
	d.keys["nil"] = (*ecdsa.PrivateKey)(nil)
	if err := d.Delete(context.Background(), "nil"); err == nil {
		t.Fatal("typed nil key was deleted as a live key")
	}
}

func TestDevHSMReinitializesAfterDestroy(t *testing.T) {
	d := NewSignerFactory(true)
	d.Destroy()
	if _, _, err := d.Generate(context.Background(), "ECDSA-P256", "after-destroy"); err != nil {
		t.Fatalf("destroy prevented reinitialization: %v", err)
	}
}

func TestDisabledHSMRejectsCapabilities(t *testing.T) {
	if _, err := NewSignerFactory(false).Capabilities(context.Background()); err == nil {
		t.Fatal("disabled HSM reported capabilities")
	}
}
