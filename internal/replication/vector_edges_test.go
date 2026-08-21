package replication

import "testing"

func TestVectorCloneReturnsEmptyIndependentMap(t *testing.T) {
	v := Vector{}
	c := v.Clone()
	c["new"] = 1
	if len(v) != 0 {
		t.Fatalf("clone mutated empty input: %#v", v)
	}
}
