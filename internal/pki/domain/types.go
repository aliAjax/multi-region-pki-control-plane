package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

type ID string
type Tenant struct{ Organization, Service, Environment, Region string }

func (t Tenant) Validate() error {
	if strings.TrimSpace(t.Organization) == "" || strings.TrimSpace(t.Service) == "" || strings.TrimSpace(t.Environment) == "" || strings.TrimSpace(t.Region) == "" {
		return errors.New("all tenant dimensions are required")
	}
	return nil
}
func (t Tenant) Key() string {
	return strings.Join([]string{t.Organization, t.Service, t.Environment, t.Region}, "/")
}

type KeyAlgorithm string

const (
	RSA2048   KeyAlgorithm = "RSA-2048"
	RSA3072   KeyAlgorithm = "RSA-3072"
	ECDSAP256 KeyAlgorithm = "ECDSA-P256"
	ECDSAP384 KeyAlgorithm = "ECDSA-P384"
	Ed25519   KeyAlgorithm = "ED25519"
)

func (k KeyAlgorithm) Secure() bool {
	return k == RSA2048 || k == RSA3072 || k == ECDSAP256 || k == ECDSAP384 || k == Ed25519
}

type CAType string

const (
	RootCA         CAType = "root"
	IntermediateCA CAType = "intermediate"
)

type CAStatus string

const (
	CAPending  CAStatus = "pending"
	CAActive   CAStatus = "active"
	CARotating CAStatus = "rotating"
	CARetired  CAStatus = "retired"
	CARevoked  CAStatus = "revoked"
)

type CertificateAuthority struct {
	ID             ID           `json:"id"`
	Name           string       `json:"name"`
	Type           CAType       `json:"type"`
	ParentID       ID           `json:"parent_id,omitempty"`
	Tenant         Tenant       `json:"tenant"`
	Region         string       `json:"region"`
	Algorithm      KeyAlgorithm `json:"algorithm"`
	KeyReference   string       `json:"key_reference"`
	Status         CAStatus     `json:"status"`
	CertificatePEM string       `json:"certificate_pem,omitempty"`
	NotBefore      time.Time    `json:"not_before"`
	NotAfter       time.Time    `json:"not_after"`
	Version        int64        `json:"version"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

func (c CertificateAuthority) Validate() error {
	if c.ID == "" || c.Name == "" {
		return errors.New("ca id and name required")
	}
	if !c.Algorithm.Secure() {
		return errors.New("unsupported or insecure key algorithm")
	}
	if c.KeyReference == "" {
		return errors.New("external key reference required")
	}
	if c.Type == IntermediateCA && c.ParentID == "" {
		return errors.New("intermediate requires parent")
	}
	if c.Type == RootCA && c.ParentID != "" {
		return errors.New("root cannot have parent")
	}
	if !c.NotAfter.After(c.NotBefore) {
		return errors.New("invalid ca validity")
	}
	return c.Tenant.Validate()
}
func (c *CertificateAuthority) Transition(next CAStatus, now time.Time) error {
	allowed := map[CAStatus]map[CAStatus]bool{CAPending: {CAActive: true, CARevoked: true}, CAActive: {CARotating: true, CARetired: true, CARevoked: true}, CARotating: {CAActive: true, CARetired: true, CARevoked: true}, CARetired: {CARevoked: true}}
	if !allowed[c.Status][next] {
		return fmt.Errorf("invalid ca transition %s to %s", c.Status, next)
	}
	c.Status = next
	c.Version++
	c.UpdatedAt = now
	return nil
}

type ProfileStatus string

const (
	ProfileDraft   ProfileStatus = "draft"
	ProfileActive  ProfileStatus = "active"
	ProfileRetired ProfileStatus = "retired"
)

type CertificateProfile struct {
	ID                 ID             `json:"id"`
	CAID               ID             `json:"ca_id"`
	Name               string         `json:"name"`
	Tenant             Tenant         `json:"tenant"`
	Algorithms         []KeyAlgorithm `json:"algorithms"`
	MaxValidity        time.Duration  `json:"-"`
	MaxValiditySeconds int64          `json:"max_validity_seconds"`
	DNSPatterns        []string       `json:"dns_patterns"`
	URIPrefixes        []string       `json:"uri_prefixes"`
	AllowIP            bool           `json:"allow_ip"`
	KeyUsages          []string       `json:"key_usages"`
	RequireApproval    bool           `json:"require_approval"`
	RenewalWindow      time.Duration  `json:"-"`
	Status             ProfileStatus  `json:"status"`
	Version            int64          `json:"version"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

func (p *CertificateProfile) Normalize() {
	if p.MaxValidity == 0 {
		p.MaxValidity = time.Duration(p.MaxValiditySeconds) * time.Second
	}
	if p.MaxValiditySeconds == 0 {
		p.MaxValiditySeconds = int64(p.MaxValidity / time.Second)
	}
}
func (p CertificateProfile) Validate() error {
	if p.ID == "" || p.CAID == "" || p.Name == "" {
		return errors.New("profile id, ca and name required")
	}
	if p.MaxValidity <= 0 || p.MaxValidity > 397*24*time.Hour {
		return errors.New("invalid profile validity")
	}
	if len(p.Algorithms) == 0 {
		return errors.New("at least one algorithm required")
	}
	for _, a := range p.Algorithms {
		if !a.Secure() {
			return errors.New("insecure algorithm")
		}
	}
	return p.Tenant.Validate()
}
func (p CertificateProfile) PermitsAlgorithm(a KeyAlgorithm) bool {
	for _, v := range p.Algorithms {
		if v == a {
			return true
		}
	}
	return false
}
func (p CertificateProfile) ValidateSANs(dns []string, ips []net.IP, uris []*url.URL) error {
	if len(ips) > 0 && !p.AllowIP {
		return errors.New("ip san forbidden")
	}
	for _, name := range dns {
		ok := false
		for _, pat := range p.DNSPatterns {
			if matchDNS(pat, name) {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("dns san %q forbidden", name)
		}
	}
	for _, u := range uris {
		ok := false
		for _, pre := range p.URIPrefixes {
			if strings.HasPrefix(u.String(), pre) {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("uri san %q forbidden", u)
		}
	}
	return nil
}
func matchDNS(pattern, name string) bool {
	pattern = strings.ToLower(strings.TrimSuffix(pattern, "."))
	name = strings.ToLower(strings.TrimSuffix(name, "."))
	if strings.HasPrefix(pattern, "*.") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(name, suffix) && strings.Count(name, ".") == strings.Count(pattern, ".")
	}
	return pattern == name
}

type Fingerprint string

func NewFingerprint(der []byte) Fingerprint {
	s := sha256.Sum256(der)
	return Fingerprint(hex.EncodeToString(s[:]))
}

type ValidityWindow struct{ NotBefore, NotAfter time.Time }

func (v ValidityWindow) Validate(max time.Duration, now time.Time) error {
	if v.NotBefore.Before(now.Add(-5 * time.Minute)) {
		return errors.New("not_before too old")
	}
	if !v.NotAfter.After(v.NotBefore) {
		return errors.New("not_after before not_before")
	}
	if v.NotAfter.Sub(v.NotBefore) > max {
		return errors.New("validity exceeds profile")
	}
	return nil
}
