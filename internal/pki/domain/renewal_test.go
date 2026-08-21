package domain

import (
	"testing"
	"time"
)

func TestExpiredPublishedIsNotRenewable(t *testing.T) {
	now := time.Now()
	c := Certificate{Status: CertPublished, Validity: ValidityWindow{NotAfter: now.Add(-time.Hour)}}
	if c.Renewable(now, 24*time.Hour) {
		t.Fatal("expired certificate renewable")
	}
}
