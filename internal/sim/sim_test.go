package sim_test

import (
	"context"
	"testing"

	"github.com/BryanDGuy/halo/internal/grid"
	"github.com/BryanDGuy/halo/internal/ref"
	"github.com/BryanDGuy/halo/internal/sim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	g := grid.New(rows, cols)
	g.Set(rows/2, cols/2, 1.0)

	want := ref.Solve(g, alpha, dt, h, steps)

	tiles := grid.Decompose(g, pw)
	s := sim.New(tiles, alpha, dt, h)
	ctx := context.Background()
	s.Start(ctx)
	for range steps {
		require.True(t, s.Step(ctx))
	}
	got := grid.New(rows, cols)
	s.Collect(got)

	for i := range got.Data {
		r, c := i/cols, i%cols
		assert.InDeltaf(t, want.Data[i], got.Data[i], 1e-10, "cell (%d,%d)", r, c)
	}
}

func TestStepReturnsFalseOnCancel(t *testing.T) {
	g := grid.New(4, 4)
	tiles := grid.Decompose(g, 1)
	s := sim.New(tiles, 0.1, 0.001, 0.25)
	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)
	cancel()
	assert.False(t, s.Step(ctx), "Step should return false after context cancel")
}
