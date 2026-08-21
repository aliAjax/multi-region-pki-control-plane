package domain

import (
	"errors"
	"strings"
	"time"
)

type AccountStatus string

const (
	AccountValid       AccountStatus = "valid"
	AccountDeactivated AccountStatus = "deactivated"
	AccountRevoked     AccountStatus = "revoked"
)

type Account struct {
	ID            string        `json:"id"`
	TenantKey     string        `json:"tenant_key"`
	KeyThumbprint string        `json:"key_thumbprint"`
	Contacts      []string      `json:"contacts"`
	TermsAgreed   bool          `json:"terms_agreed"`
	Status        AccountStatus `json:"status"`
	Version       int64         `json:"version"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

func (a Account) Validate() error {
	if a.ID == "" || a.TenantKey == "" || a.KeyThumbprint == "" {
		return errors.New("account identity required")
	}
	for _, c := range a.Contacts {
		if !strings.HasPrefix(c, "mailto:") {
			return errors.New("only mailto contact supported")
		}
	}
	return nil
}
func (a *Account) RotateKey(old, new string, now time.Time) error {
	if a.Status != AccountValid {
		return errors.New("account inactive")
	}
	if old != a.KeyThumbprint {
		return errors.New("old key mismatch")
	}
	if new == "" || new == old {
		return errors.New("invalid new key")
	}
	a.KeyThumbprint = new
	a.Version++
	a.UpdatedAt = now
	return nil
}
func (a *Account) Deactivate(now time.Time) error {
	if a.Status != AccountValid {
		return errors.New("account not active")
	}
	a.Status = AccountDeactivated
	a.Version++
	a.UpdatedAt = now
	return nil
}
