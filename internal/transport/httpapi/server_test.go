package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/pki-control-plane/internal/config"
	"example.com/pki-control-plane/internal/crypto"
	"example.com/pki-control-plane/internal/issuance"
	"example.com/pki-control-plane/internal/repository"
	"example.com/pki-control-plane/internal/revocation"
)

func TestRevokeEndpointReturnsNotFoundForMissingCertificate(t *testing.T) {
	s := repository.NewMemoryStore()
	hsm := crypto.NewSignerFactory(false)
	h := NewServer(config.Config{Region: "test"}, slog.Default(), s, issuance.NewService(s, hsm), revocation.NewService(s), hsm)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/certificates/missing/revoke", strings.NewReader(`{"reason":"key_compromise"}`))
	r = r.WithContext(context.Background())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing certificate mapped to %d: %s", w.Code, w.Body.String())
	}
}
