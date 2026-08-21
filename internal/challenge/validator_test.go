package challenge

import (
	"context"
	"sync"
	"testing"

	acme "example.com/pki-control-plane/internal/acme/domain"
)

type immediateValidator struct{}
func (immediateValidator) Validate(context.Context, Request) (Result, error) { return Result{Valid: true}, nil }

func TestRegistryConcurrentRegisterValidate(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) { defer wg.Done(); <-start; _ = r.Register(acme.ChallengeType("custom"+string(rune('a'+i))), immediateValidator{}) }(i)
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) { defer wg.Done(); <-start; _, _ = r.Validate(context.Background(), Request{Type: acme.ChallengeType("custom"+string(rune('a'+i)))}) }(i)
	}
	close(start)
	wg.Wait()
}
