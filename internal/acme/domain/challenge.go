package domain

import (
	"errors"
	"fmt"
	"time"
)

type AuthorizationStatus string

const (
	AuthzPending     AuthorizationStatus = "pending"
	AuthzValid       AuthorizationStatus = "valid"
	AuthzInvalid     AuthorizationStatus = "invalid"
	AuthzDeactivated AuthorizationStatus = "deactivated"
	AuthzExpired     AuthorizationStatus = "expired"
	AuthzRevoked     AuthorizationStatus = "revoked"
)

type ChallengeType string

const (
	HTTP01    ChallengeType = "http-01"
	DNS01     ChallengeType = "dns-01"
	TLSALPN01 ChallengeType = "tls-alpn-01"
)

type ChallengeStatus string

const (
	ChallengePending    ChallengeStatus = "pending"
	ChallengeProcessing ChallengeStatus = "processing"
	ChallengeValid      ChallengeStatus = "valid"
	ChallengeInvalid    ChallengeStatus = "invalid"
)

type Challenge struct {
	ID              string          `json:"id"`
	AuthorizationID string          `json:"authorization_id"`
	Type            ChallengeType   `json:"type"`
	Token           string          `json:"token"`
	Status          ChallengeStatus `json:"status"`
	ValidatedAt     *time.Time      `json:"validated_at,omitempty"`
	Error           string          `json:"error,omitempty"`
	ExpiresAt       time.Time       `json:"expires_at"`
	Attempts        int             `json:"attempts"`
	Version         int64           `json:"version"`
}

func (c Challenge) Validate(now time.Time) error {
	if c.ID == "" || c.AuthorizationID == "" || c.Token == "" {
		return errors.New("challenge identity required")
	}
	if c.Type != HTTP01 && c.Type != DNS01 && c.Type != TLSALPN01 {
		return errors.New("unknown challenge type")
	}
	if !c.ExpiresAt.After(now) {
		return errors.New("challenge expired")
	}
	return nil
}
func (c *Challenge) Start(now time.Time) error {
	if c.Status != ChallengePending {
		return fmt.Errorf("challenge not pending: %s", c.Status)
	}
	if !now.Before(c.ExpiresAt) {
		c.Status = ChallengeInvalid
		c.Error = "expired"
		return errors.New("challenge expired")
	}
	c.Status = ChallengeProcessing
	c.Attempts++
	c.Version++
	return nil
}
func (c *Challenge) Complete(ok bool, detail string, now time.Time) error {
	if c.Status != ChallengeProcessing {
		return errors.New("challenge not processing")
	}
	if ok {
		c.Status = ChallengeValid
		c.ValidatedAt = &now
	} else {
		c.Status = ChallengeInvalid
		c.Error = detail
	}
	c.Version++
	return nil
}

type Authorization struct {
	ID           string              `json:"id"`
	AccountID    string              `json:"account_id"`
	Identifier   Identifier          `json:"identifier"`
	Wildcard     bool                `json:"wildcard"`
	ChallengeIDs []string            `json:"challenge_ids"`
	Status       AuthorizationStatus `json:"status"`
	ExpiresAt    time.Time           `json:"expires_at"`
	Version      int64               `json:"version"`
}

func (a *Authorization) Apply(c Challenge, now time.Time) error {
	if c.AuthorizationID != a.ID {
		return errors.New("challenge belongs to another authorization")
	}
	if c.Status == ChallengeValid {
		a.Status = AuthzValid
		a.Version++
		return nil
	}
	if c.Status == ChallengeInvalid && now.After(a.ExpiresAt) {
		a.Status = AuthzExpired
		a.Version++
	}
	return nil
}
