package crypto

import (
	"errors"
	"testing"
)

func TestParseCertificateClassifiesInvalidPEM(t *testing.T) {
	_, e := ParseCertificate("bad")
	if !errors.Is(e, ErrInvalidCertificatePEM) {
		t.Fatalf("lost: %v", e)
	}
}
