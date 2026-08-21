package errors

import (
	stderrors "errors"
	"testing"
)

func TestPublicErrorPreservesCauseChain(t *testing.T) {
	c := stderrors.New("missing")
	e := Wrap(NotFound, "load", c)
	if !stderrors.Is(e, c) {
		t.Fatalf("lost: %v", e)
	}
}
