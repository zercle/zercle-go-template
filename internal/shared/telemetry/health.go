// Health registry for liveness and readiness probes.
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Checker is a dependency health probe.
type Checker interface {
	Name() string
	Check(ctx context.Context) error
}

// Registry holds liveness and readiness checkers. Liveness only needs the
// process to be alive; readiness reflects the health of real dependencies.
type Registry struct {
	livenessMu  sync.RWMutex
	readinessMu sync.RWMutex
	liveness    []Checker
	readiness   []Checker
}

// NewRegistry returns an empty health registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// AddLiveness registers a checker that only verifies the process is alive.
func (r *Registry) AddLiveness(c Checker) {
	r.livenessMu.Lock()
	defer r.livenessMu.Unlock()
	r.liveness = append(r.liveness, c)
}

// AddReadiness registers a checker that verifies a dependency is healthy.
func (r *Registry) AddReadiness(c Checker) {
	r.readinessMu.Lock()
	defer r.readinessMu.Unlock()
	r.readiness = append(r.readiness, c)
}

// Live always returns nil because process liveness is handled by the endpoint
// itself responding. Custom liveness checkers may be added for completeness.
func (r *Registry) Live(ctx context.Context) error {
	r.livenessMu.RLock()
	checkers := append([]Checker(nil), r.liveness...)
	r.livenessMu.RUnlock()

	return runCheckers(ctx, checkers)
}

// Ready runs all readiness checkers concurrently and returns an aggregated error
// naming every failing checker.
func (r *Registry) Ready(ctx context.Context) error {
	r.readinessMu.RLock()
	checkers := append([]Checker(nil), r.readiness...)
	r.readinessMu.RUnlock()

	return runCheckers(ctx, checkers)
}

// runCheckers executes every checker concurrently and aggregates failures.
func runCheckers(ctx context.Context, checkers []Checker) error {
	if len(checkers) == 0 {
		return nil
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	wg.Add(len(checkers))
	for _, c := range checkers {
		go func(checker Checker) {
			defer wg.Done()

			if err := checker.Check(ctx); err != nil {
				mu.Lock()
				defer mu.Unlock()
				errs = append(errs, fmt.Errorf("%s: %w", checker.Name(), err))
			}
		}(c)
	}
	wg.Wait()

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}
