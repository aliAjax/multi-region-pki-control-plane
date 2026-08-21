package domain

import (
	"time"
)

func (c Certificate) Renewable(now time.Time, window time.Duration) bool {
	return c.Status == CertPublished && now.Add(window).After(c.Validity.NotAfter)
}

func (c Certificate) ActiveAt(t time.Time) bool {
	return c.Status == CertIssued || c.Status == CertPublished || c.Status == CertRenewing && !t.Before(c.Validity.NotBefore) && t.Before(c.Validity.NotAfter)
}
