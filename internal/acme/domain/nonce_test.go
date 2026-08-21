package domain

import (
	"testing"
	"time"
)

func TestNonceExpiryAndReplayAreAtomic(t *testing.T) {
	now := time.Unix(100, 0)
	n := NewNonceStore(time.Second, 100)
	v, err := n.Issue(now)
	if err != nil { t.Fatal(err) }
	if err := n.Consume(v, now.Add(500*time.Millisecond)); err != nil { t.Fatal(err) }
	if err := n.Consume(v, now.Add(500*time.Millisecond)); err == nil { t.Fatal("replay accepted") }
	v2, err := n.Issue(now.Add(2 * time.Second))
	if err != nil { t.Fatal(err) }
	if err := n.Consume(v2, now.Add(4*time.Second)); err == nil { t.Fatal("expired nonce accepted") }
}
