package replication

import (
	"errors"
	"sort"
)

type Vector map[string]uint64
type Relation string

const (
	Equal      Relation = "equal"
	Before     Relation = "before"
	After      Relation = "after"
	Concurrent Relation = "concurrent"
)

func (v Vector) Increment(region string) Vector { r := v.Clone(); r[region]++; return r }
func (v Vector) Clone() Vector {
	r := make(Vector, len(v))
	for k, n := range v {
		r[k] = n
	}
	return r
}
func (v Vector) Compare(o Vector) Relation {
	less, greater := false, false
	keys := map[string]struct{}{}
	for k := range v {
		keys[k] = struct{}{}
	}
	for k := range o {
		keys[k] = struct{}{}
	}
	for k := range keys {
		if v[k] < o[k] {
			less = true
		}
		if v[k] > o[k] {
			greater = true
		}
	}
	if less && greater {
		return Concurrent
	}
	if less {
		return Before
	}
	if greater {
		return After
	}
	return Equal
}
func (v Vector) Merge(o Vector) Vector {
	r := v.Clone()
	for k, n := range o {
		if n > r[k] {
			r[k] = n
		}
	}
	return r
}
func (v Vector) Regions() []string {
	r := make([]string, 0, len(v))
	for k := range v {
		r = append(r, k)
	}
	sort.Strings(r)
	return r
}

type Envelope struct {
	AggregateID  string `json:"aggregate_id"`
	Kind         string `json:"kind"`
	Region       string `json:"region"`
	Vector       Vector `json:"vector"`
	FencingToken uint64 `json:"fencing_token"`
	Payload      []byte `json:"payload"`
	Checksum     string `json:"checksum"`
}

func (e Envelope) Validate() error {
	if e.AggregateID == "" || e.Kind == "" || e.Region == "" {
		return errors.New("replication identity required")
	}
	if e.FencingToken == 0 {
		return errors.New("fencing token required")
	}
	if len(e.Payload) == 0 {
		return errors.New("payload required")
	}
	return nil
}
