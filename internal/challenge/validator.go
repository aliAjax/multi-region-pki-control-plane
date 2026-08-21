package challenge

import (
	"context"
	"errors"
	acme "example.com/pki-control-plane/internal/acme/domain"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Request struct {
	Type             acme.ChallengeType
	Identifier       string
	Token            string
	KeyAuthorization string
}
type Result struct {
	Valid      bool
	Detail     string
	ObservedAt time.Time
}
type Validator interface {
	Validate(context.Context, Request) (Result, error)
}
type Registry struct {
	mu         sync.RWMutex
	validators map[acme.ChallengeType]Validator
}

func NewRegistry() *Registry { return &Registry{validators: map[acme.ChallengeType]Validator{}} }
func (r *Registry) Register(t acme.ChallengeType, v Validator) error {
	if v == nil {
		return errors.New("nil validator")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.validators[t]; ok {
		return errors.New("validator already registered")
	}
	r.validators[t] = v
	return nil
}
func (r *Registry) Validate(ctx context.Context, q Request) (Result, error) {
	r.mu.RLock()
	v, ok := r.validators[q.Type]
	r.mu.RUnlock()
	if !ok {
		return Result{}, fmt.Errorf("no validator for %s", q.Type)
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return v.Validate(ctx, q)
}

type HTTPValidator struct{ Client *http.Client }

func (v HTTPValidator) Validate(ctx context.Context, q Request) (Result, error) {
	client := v.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	u := "http://" + q.Identifier + "/.well-known/acme-challenge/" + q.Token
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return Result{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Result{Detail: resp.Status, ObservedAt: time.Now().UTC()}, nil
	}
	buf := make([]byte, len(q.KeyAuthorization)+1)
	n, _ := resp.Body.Read(buf)
	valid := strings.TrimSpace(string(buf[:n])) == q.KeyAuthorization
	return Result{Valid: valid, Detail: "http-01", ObservedAt: time.Now().UTC()}, nil
}

type DNSResolver interface {
	LookupTXT(context.Context, string) ([]string, error)
}
type netResolver struct{ *net.Resolver }

func (n netResolver) LookupTXT(ctx context.Context, name string) ([]string, error) {
	return n.Resolver.LookupTXT(ctx, name)
}

type DNSValidator struct{ Resolver DNSResolver }

func (v DNSValidator) Validate(ctx context.Context, q Request) (Result, error) {
	r := v.Resolver
	if r == nil {
		r = netResolver{net.DefaultResolver}
	}
	records, err := r.LookupTXT(ctx, "_acme-challenge."+q.Identifier)
	if err != nil {
		return Result{}, err
	}
	for _, record := range records {
		if record == q.KeyAuthorization {
			return Result{Valid: true, Detail: "dns-01", ObservedAt: time.Now().UTC()}, nil
		}
	}
	return Result{Detail: "TXT value not found", ObservedAt: time.Now().UTC()}, nil
}

type TLSALPNValidator struct{}

func (TLSALPNValidator) Validate(context.Context, Request) (Result, error) {
	return Result{}, errors.New("tls-alpn-01 network adapter unimplemented")
}
