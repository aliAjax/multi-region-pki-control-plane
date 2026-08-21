package domain

import (
	"errors"
	"fmt"
	"time"
)

type OrderStatus string

const (
	OrderPending    OrderStatus = "pending"
	OrderReady      OrderStatus = "ready"
	OrderProcessing OrderStatus = "processing"
	OrderValid      OrderStatus = "valid"
	OrderInvalid    OrderStatus = "invalid"
)

type Identifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
type Order struct {
	ID               string       `json:"id"`
	AccountID        string       `json:"account_id"`
	Identifiers      []Identifier `json:"identifiers"`
	AuthorizationIDs []string     `json:"authorization_ids"`
	CertificateID    string       `json:"certificate_id,omitempty"`
	Status           OrderStatus  `json:"status"`
	ExpiresAt        time.Time    `json:"expires_at"`
	NotBefore        time.Time    `json:"not_before"`
	NotAfter         time.Time    `json:"not_after"`
	Error            string       `json:"error,omitempty"`
	IdempotencyKey   string       `json:"idempotency_key"`
	Version          int64        `json:"version"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

func (o Order) Validate(now time.Time) error {
	if o.ID == "" || o.AccountID == "" || o.IdempotencyKey == "" {
		return errors.New("order identity required")
	}
	if len(o.Identifiers) == 0 {
		return errors.New("identifiers required")
	}
	if !o.ExpiresAt.After(now) {
		return errors.New("order already expired")
	}
	for _, i := range o.Identifiers {
		if i.Type != "dns" || i.Value == "" {
			return errors.New("only non-empty dns identifier supported")
		}
	}
	return nil
}
func (o *Order) Transition(next OrderStatus, now time.Time) error {
	allowed := map[OrderStatus]map[OrderStatus]bool{OrderPending: {OrderReady: true, OrderInvalid: true}, OrderReady: {OrderProcessing: true, OrderInvalid: true}, OrderProcessing: {OrderValid: true, OrderInvalid: true}}
	if !allowed[o.Status][next] {
		return fmt.Errorf("invalid order transition %s to %s", o.Status, next)
	}
	o.Status = next
	o.Version++
	o.UpdatedAt = now
	return nil
}
func (o *Order) Expire(now time.Time) bool {
	if now.Before(o.ExpiresAt) || o.Status == OrderValid || o.Status == OrderInvalid {
		return false
	}
	o.Status = OrderInvalid
	o.Error = "order expired"
	o.Version++
	o.UpdatedAt = now
	return true
}
