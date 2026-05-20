// internal/sim/sim_test.go
package sim_test

import (
	"context"
	"math"
	"testing"

	"github.com/BryanDGuy/halo/internal/grid"
	"github.com/BryanDGuy/halo/internal/ref"
	"github.com/BryanDGuy/halo/internal/sim"
)

func TestSimMatchesRef(t *testing.T) {
	const (
		rows  = 16
		cols  = 16
		pw    = 2
		steps = 50
	)
	alpha := 0.1
	h := 1.0 / float64(max(rows, cols))
	dt := 0.24 * h * h / alpha

	// Shared initial condition: hotspot at center
	g := grid.New(rows, cols)
	g.Set(rows/2, cols/2, 1.0)

	// Reference result
	want := ref.Solve(g, alpha, dt, h, steps)

	// Concurrent result
	tiles := grid.Decompose(g, pw)
	s := sim.New(tiles, alpha, dt, h)
	ctx := context.Background()
	s.Start(ctx)
	for i := 0; i < steps; i++ {
		s.Step(ctx)
	}
	got := grid.New(rows, cols)
	s.Collect(got)

	for i := range got.Data {
		if math.Abs(got.Data[i]-want.Data[i]) > 1e-10 {
			r, c := i/cols, i%cols
			t.Errorf("cell (%d,%d): want %.15f, got %.15f", r, c, want.Data[i], got.Data[i])
		}
	}
}

func TestStepReturnsFalseOnCancel(t *testing.T) {
	g := grid.New(4, 4)
	tiles := grid.Decompose(g, 1)
	s := sim.New(tiles, 0.1, 0.001, 0.25)
	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)
	cancel()
	// After cancel, Step should return false
	if s.Step(ctx) {
		t.Error("Step should return false after context cancel")
	}
}
