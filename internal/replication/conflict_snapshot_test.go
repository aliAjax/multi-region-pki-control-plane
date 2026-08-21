package replication

import (
	"context"
	"testing"
)

func TestSyncerConflictsReturnsDeepSnapshots(t *testing.T) {
	s := NewSyncer(acceptApply{})
	_ = s.Sync(context.Background(), Envelope{AggregateID: "a", Kind: "k", Region: "tokyo", Vector: Vector{"tokyo": 1}, FencingToken: 1, Payload: []byte("x")})
	_ = s.Sync(context.Background(), Envelope{AggregateID: "a", Kind: "k", Region: "sg", Vector: Vector{"sg": 1}, FencingToken: 2, Payload: []byte("y")})
	c := s.Conflicts()
	if len(c) != 1 {
		t.Fatalf("conflicts: %d", len(c))
	}
	c[0].Remote.Vector["sg"] = 99
	if s.Conflicts()[0].Remote.Vector["sg"] != 1 {
		t.Fatal("conflict snapshot leaked internal vector")
	}
}
