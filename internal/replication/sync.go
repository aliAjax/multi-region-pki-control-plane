package replication

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Conflict struct {
	Local, Remote Envelope
	Reason        string
	DetectedAt    time.Time
}
type Applier interface {
	Apply(context.Context, Envelope) error
}
type Syncer struct {
	mu        sync.Mutex
	vectors   map[string]Vector
	fences    map[string]uint64
	conflicts []Conflict
	apply     Applier
}

func NewSyncer(a Applier) *Syncer {
	return &Syncer{vectors: map[string]Vector{}, fences: map[string]uint64{}, apply: a}
}
func (s *Syncer) Sync(ctx context.Context, e Envelope) error {
	if err := e.Validate(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if e.FencingToken < s.fences[e.AggregateID] {
		return errors.New("stale fencing token")
	}
	local := s.vectors[e.AggregateID]
	relation := local.Compare(e.Vector)
	if relation == After || relation == Equal {
		return nil
	}
	if relation == Concurrent {
		s.conflicts = append(s.conflicts, Conflict{Remote: e, Reason: "concurrent metadata update", DetectedAt: time.Now().UTC()})
		return errors.New("replication conflict")
	}
	if err := s.apply.Apply(ctx, e); err != nil {
		return err
	}
	s.fences[e.AggregateID] = e.FencingToken
	s.vectors[e.AggregateID] = local.Merge(e.Vector)
	return nil
}
func (s *Syncer) Cursor(id string) Vector {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.vectors[id]
}
func (s *Syncer) Conflicts() []Conflict {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Conflict(nil), s.conflicts...)
}
