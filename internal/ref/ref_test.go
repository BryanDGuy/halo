package ref_test

import (
	"testing"

	"github.com/BryanDGuy/halo/internal/grid"
	"github.com/BryanDGuy/halo/internal/ref"
	"github.com/stretchr/testify/assert"
)

func TestRefOneStep(t *testing.T) {
	// 3x3 grid, center=1.0, all else 0.
	// alpha=1, dt=0.1, h=1 → r=0.1 (stable: 0.1 ≤ 0.25)
	// After one step:
	//   center (1,1): 1.0 + 0.1*(0+0+0+0 - 4*1.0) = 0.6
	//   boundary cells: 0 (fixed, never updated)
	g := grid.New(3, 3)
	g.Set(1, 1, 1.0)
	result := ref.Solve(g, 1.0, 0.1, 1.0, 1)

	assert.InDelta(t, 0.6, result.At(1, 1), 1e-14)
	assert.InDelta(t, 0.0, result.At(0, 1), 1e-14, "boundary row must stay 0")
	assert.InDelta(t, 0.0, result.At(1, 0), 1e-14, "boundary col must stay 0")
}

func TestRefBoundariesFixed(t *testing.T) {
	g := grid.New(4, 4)
	g.Set(2, 2, 1.0)
	result := ref.Solve(g, 0.1, 0.001, 0.25, 100)
	for c := range 4 {
		assert.InDelta(t, 0.0, result.At(0, c), 1e-14, "top boundary col %d", c)
		assert.InDelta(t, 0.0, result.At(3, c), 1e-14, "bottom boundary col %d", c)
	}
	for r := range 4 {
		assert.InDelta(t, 0.0, result.At(r, 0), 1e-14, "left boundary row %d", r)
		assert.InDelta(t, 0.0, result.At(r, 3), 1e-14, "right boundary row %d", r)
	}
}
