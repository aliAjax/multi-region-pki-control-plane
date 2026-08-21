package replication

import "testing"

func TestVectorMergeDoesNotMutateInputs(t *testing.T) {
	left := Vector{"tokyo": 2}
	right := Vector{"singapore": 4}
	merged := left.Merge(right)
	merged["tokyo"] = 99
	merged["new"] = 1
	if left["tokyo"] != 2 || len(left) != 1 {
		t.Fatalf("left vector was contaminated: %#v", left)
	}
	if right["singapore"] != 4 || len(right) != 1 {
		t.Fatalf("right vector was contaminated: %#v", right)
	}
}
