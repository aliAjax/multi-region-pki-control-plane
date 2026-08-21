package crypto

import (
	"context"
	stdcrypto "crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"errors"
	"fmt"
	"sync"
)

type DevHSM struct {
	mu      sync.RWMutex
	keys    map[string]stdcrypto.Signer
	enabled bool
}

func NewSignerFactory(enabled bool) *DevHSM {
	return &DevHSM{keys: make(map[string]stdcrypto.Signer), enabled: enabled}
}
func (d *DevHSM) Generate(ctx context.Context, algorithm, label string) (string, stdcrypto.PublicKey, error) {
	if !d.enabled {
		return "", nil, errors.New("development HSM disabled; configure production HSM/KMS")
	}
	select {
	case <-ctx.Done():
		return "", nil, ctx.Err()
	default:
	}
	if algorithm != "ECDSA-P256" {
		return "", nil, fmt.Errorf("dev HSM supports only ECDSA-P256, got %s", algorithm)
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", nil, err
	}
	ref := "devhsm://" + label
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.keys[ref]; ok {
		return "", nil, errors.New("key label exists")
	}
	d.keys[ref] = key
	return ref, key.Public(), nil
}
func (d *DevHSM) Signer(ctx context.Context, ref string) (stdcrypto.Signer, error) {
	if !d.enabled {
		return nil, errors.New("development HSM disabled")
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	s, ok := d.keys[ref]
	if !ok || signerIsNil(s) {
		return nil, errors.New("HSM key not found")
	}
	return s, nil
}
func (d *DevHSM) Available(ctx context.Context, ref string) error {
	_, err := d.Signer(ctx, ref)
	return err
}
func (d *DevHSM) Delete(ctx context.Context, ref string) error {
	_ = ctx
	d.mu.Lock()
	defer d.mu.Unlock()
	s, ok := d.keys[ref]
	if !ok || signerIsNil(s) {
		return errors.New("key not found")
	}
	delete(d.keys, ref)
	return nil
}
func (d *DevHSM) Capabilities(context.Context) (Capabilities, error) {
	if !d.enabled {
		return Capabilities{}, errors.New("development HSM disabled")
	}
	return Capabilities{Algorithms: []string{"ECDSA-P256"}, HardwareBacked: false, Attestation: false}, nil
}
func (d *DevHSM) Destroy() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for k := range d.keys {
		delete(d.keys, k)
	}
	d.keys = make(map[string]stdcrypto.Signer)
}
func (d *DevHSM) Sign(ctx context.Context, tpl, parent *x509.Certificate, pub any, keyRef string) ([]byte, error) {
	s, err := d.Signer(ctx, keyRef)
	if err != nil {
		return nil, err
	}
	done := make(chan struct {
		der []byte
		err error
	}, 1)
	go func() {
		b, e := x509.CreateCertificate(rand.Reader, tpl, parent, pub, s)
		done <- struct {
			der []byte
			err error
		}{b, e}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-done:
		return r.der, r.err
	}
}
