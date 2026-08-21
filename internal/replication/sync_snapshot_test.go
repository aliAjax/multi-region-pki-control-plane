package replication

import (
	"context"
	"testing"
)

type acceptApply struct{}

func (acceptApply) Apply(context.Context, Envelope) error { return nil }
func TestSyncerCursorDoesNotExposeInternalVector(t *testing.T) {
	s := NewSyncer(acceptApply{})
	e := Envelope{AggregateID: "a", Kind: "k", Region: "r", Vector: Vector{"r": 1}, FencingToken: 1, Payload: []byte("x")}
	if s.Sync(context.Background(), e) != nil {
		t.Fatal("sync")
	}
	c := s.Cursor("a")
	c["r"] = 9
	if s.Cursor("a")["r"] != 1 {
		t.Fatal("alias")
	}
}
