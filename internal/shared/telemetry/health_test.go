//go:build unit
// +build unit

package telemetry_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/zercle/zercle-go-template/internal/shared/telemetry"
)

type staticChecker struct {
	name string
	err  error
}

func (c *staticChecker) Name() string                  { return c.name }
func (c *staticChecker) Check(_ context.Context) error { return c.err }

func TestRegistry_Ready_NoCheckers(t *testing.T) {
	r := telemetry.NewRegistry()
	if err := r.Ready(context.Background()); err != nil {
		t.Fatalf("expected nil when no checkers, got %v", err)
	}
}

func TestRegistry_Ready_FailingCheckerNamed(t *testing.T) {
	r := telemetry.NewRegistry()
	r.AddReadiness(&staticChecker{name: "db", err: errors.New("db unreachable")})

	err := r.Ready(context.Background())
	if err == nil {
		t.Fatal("expected error from failing checker")
	}
	if !strings.Contains(err.Error(), "db") {
		t.Fatalf("expected error to name checker db, got %v", err)
	}
}
