package domain

import (
	"time"
)

func (c Certificate) Renewable(now time.Time, window time.Duration) bool {
	return c.Status == CertPublished && now.Before(c.Validity.NotAfter) && now.Add(window).After(c.Validity.NotAfter)
}

func (c Certificate) ActiveAt(t time.Time) bool {
	switch c.Status {
	case CertIssued, CertPublished, CertRenewing:
		return !t.Before(c.Validity.NotBefore) && t.Before(c.Validity.NotAfter)
	}
	return false
}
