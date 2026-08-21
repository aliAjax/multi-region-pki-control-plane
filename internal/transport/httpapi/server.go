package httpapi

import (
	"encoding/json"
	"errors"
	"example.com/pki-control-plane/internal/config"
	cryptop "example.com/pki-control-plane/internal/crypto"
	"example.com/pki-control-plane/internal/issuance"
	"example.com/pki-control-plane/internal/pki/domain"
	"example.com/pki-control-plane/internal/repository"
	"example.com/pki-control-plane/internal/revocation"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	cfg     config.Config
	logger  *slog.Logger
	store   repository.Store
	issuer  *issuance.Service
	revoker *revocation.Service
	hsm     *cryptop.DevHSM
}

func NewServer(c config.Config, l *slog.Logger, s repository.Store, i *issuance.Service, r *revocation.Service, h *cryptop.DevHSM) http.Handler {
	return (&Server{cfg: c, logger: l, store: s, issuer: i, revoker: r, hsm: h}).routes()
}
func (s *Server) routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", s.health)
	m.HandleFunc("/readyz", s.ready)
	m.HandleFunc("/api/v1/ca", s.ca)
	m.HandleFunc("/api/v1/ca/", s.caProfiles)
	m.HandleFunc("/api/v1/certificates", s.certificates)
	m.HandleFunc("/api/v1/certificates/", s.certificateAction)
	m.HandleFunc("/api/v1/ocsp/", s.ocsp)
	m.HandleFunc("/api/v1/audit/export", s.audit)
	m.HandleFunc("/acme/directory", s.directory)
	m.HandleFunc("/acme/new-nonce", s.nonce)
	return requestID(m)
}

func (s *Server) caProfiles(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/ca/"), "/")
	if len(parts) != 2 || parts[1] != "profiles" {
		writeError(w, http.StatusNotFound, "not_found", "profile endpoint not found")
		return
	}
	if r.Method == http.MethodGet {
		write(w, http.StatusOK, map[string]any{"items": []domain.CertificateProfile{}})
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	var p domain.CertificateProfile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	p.CAID = domain.ID(parts[0])
	p.Normalize()
	p.Status = domain.ProfileActive
	p.Version = 1
	p.CreatedAt = time.Now().UTC()
	p.UpdatedAt = p.CreatedAt
	if err := p.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := s.store.SaveProfile(r.Context(), &p); err != nil {
		writeError(w, http.StatusConflict, "conflict", err.Error())
		return
	}
	write(w, http.StatusCreated, p)
}
func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = time.Now().UTC().Format("20060102150405.000000")
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	write(w, http.StatusOK, map[string]any{"status": "ok", "region": s.cfg.Region, "time": time.Now().UTC()})
}
func (s *Server) ready(w http.ResponseWriter, _ *http.Request) {
	write(w, http.StatusOK, map[string]string{"status": "ready"})
}
func (s *Server) ca(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		write(w, http.StatusOK, map[string]any{"items": s.store.ListCAs(r.Context())})
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	var req struct {
		Name            string              `json:"name"`
		Type            domain.CAType       `json:"type"`
		ParentID        domain.ID           `json:"parent_id"`
		Tenant          domain.Tenant       `json:"tenant"`
		Region          string              `json:"region"`
		Algorithm       domain.KeyAlgorithm `json:"algorithm"`
		ValiditySeconds int64               `json:"validity_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid_request", err.Error())
		return
	}
	svc := issuance.NewCAService(s.store, s.hsm)
	ca, err := svc.Create(r.Context(), issuance.CreateCARequest{Name: req.Name, Type: req.Type, ParentID: req.ParentID, Tenant: req.Tenant, Region: req.Region, Algorithm: req.Algorithm, Validity: time.Duration(req.ValiditySeconds) * time.Second, Actor: "api"})
	if err != nil {
		writeError(w, 400, "invalid_request", err.Error())
		return
	}
	write(w, http.StatusCreated, ca)
}
func (s *Server) certificates(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		write(w, 200, map[string]any{"items": s.store.ListCertificates(r.Context())})
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, 405, "method_not_allowed", "method not allowed")
		return
	}
	var req issuance.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, 400, "invalid_request", err.Error())
		return
	}
	req.IdempotencyKey = r.Header.Get("Idempotency-Key")
	if req.Actor == "" {
		req.Actor = "api"
	}
	c, err := s.issuer.Request(r.Context(), req)
	if err != nil {
		writeError(w, 400, "invalid_request", err.Error())
		return
	}
	write(w, 201, c)
}
func (s *Server) certificateAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/certificates/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, 404, "not_found", "certificate not found")
		return
	}
	id := parts[0]
	if len(parts) > 1 && parts[1] == "renew" {
		writeError(w, 501, "unimplemented", "renewal worker is asynchronous")
		return
	}
	if len(parts) > 1 && parts[1] == "revoke" {
		var req struct {
			Reason revocation.Reason `json:"reason"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		c, err := s.revoker.Revoke(r.Context(), id, req.Reason, "api", 0)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				writeError(w, http.StatusNotFound, "not_found", err.Error())
				return
			}
			writeError(w, 400, "invalid_request", err.Error())
			return
		}
		write(w, 200, c)
		return
	}
	c, err := s.store.GetCertificate(r.Context(), domain.ID(id))
	if err != nil {
		writeError(w, 404, "not_found", err.Error())
		return
	}
	write(w, 200, c)
}
func (s *Server) ocsp(w http.ResponseWriter, r *http.Request) {
	serial := strings.TrimPrefix(r.URL.Path, "/api/v1/ocsp/")
	o, err := s.revoker.Check(r.Context(), serial)
	if err != nil {
		writeError(w, 500, "internal", err.Error())
		return
	}
	write(w, 200, o)
}
func (s *Server) audit(w http.ResponseWriter, r *http.Request) {
	write(w, 200, map[string]any{"items": s.store.Audits(r.Context())})
}
func (s *Server) directory(w http.ResponseWriter, _ *http.Request) {
	write(w, 200, map[string]string{"newNonce": "/acme/new-nonce", "newAccount": "/acme/new-account", "newOrder": "/acme/new-order"})
}
func (s *Server) nonce(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Replay-Nonce", time.Now().UTC().Format("20060102T150405.000000000Z"))
	w.WriteHeader(204)
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, code, msg string) {
	write(w, status, map[string]any{"error": map[string]string{"code": code, "message": msg}})
}
