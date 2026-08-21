package domain

import (
	"testing"
	"time"
)

func TestLifecycleTransitionAndActivity(t *testing.T) {
	now := time.Now().UTC()
	c := Certificate{Status: CertIssued, Validity: ValidityWindow{NotBefore: now.Add(-time.Minute), NotAfter: now.Add(time.Hour)}}
	if !c.ActiveAt(now) {
		t.Fatal("issued certificate should be active")
	}
	if c.ActiveAt(now.Add(2 * time.Hour)) {
		t.Fatal("expired issued certificate should not be active")
	}
	if err := c.Transition(CertPublished, "operator", now); err != nil {
		t.Fatal(err)
	}
	if !c.ActiveAt(now) || c.Version != 1 {
		t.Fatalf("published certificate state invalid: %#v", c)
	}
	if err := c.Transition(CertRenewing, "worker", now); err != nil {
		t.Fatal(err)
	}
	if !c.ActiveAt(now) {
		t.Fatal("renewing certificate should remain active")
	}
}
