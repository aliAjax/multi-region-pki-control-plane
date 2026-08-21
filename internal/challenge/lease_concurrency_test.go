package challenge

import (
	"testing"
	"time"
)

func TestLeaseReleaseRejectsStaleConcurrentOwner(t *testing.T) {
	m := NewLeaseManager()
	n := time.Now()
	a, _ := m.Acquire("k", "a", time.Minute, n)
	b, _ := m.Acquire("k", "b", time.Minute, n.Add(2*time.Minute))
	if m.Release(a) == nil {
		t.Fatal("stale release accepted")
	}
	if _, e := m.Renew(b, time.Minute, n.Add(150*time.Second)); e != nil {
		t.Fatal(e)
	}
}

func TestLeaseReleaseSynchronizesWithRenew(t *testing.T) {
	m := NewLeaseManager()
	now := time.Now()
	lease, _ := m.Acquire("shared", "worker", time.Minute, now)
	start := make(chan struct{})
	done := make(chan struct{}, 2)
	go func() { <-start; _, _ = m.Renew(lease, time.Minute, now.Add(time.Second)); done <- struct{}{} }()
	go func() { <-start; _ = m.Release(lease); done <- struct{}{} }()
	close(start)
	<-done
	<-done
}
