// internal/ref/ref_test.go
package ref_test

import (
	"math"
	"testing"

	"github.com/BryanDGuy/halo/internal/grid"
	"github.com/BryanDGuy/halo/internal/ref"
)

func TestRefOneStep(t *testing.T) {
	// 3x3 grid, center=1.0, all else 0.
	// alpha=1, dt=0.1, h=1 → r=0.1 (stable: 0.1 ≤ 0.25)
	// After one step:
	//   center (1,1): 1.0 + 0.1*(0+0+0+0 - 4*1.0) = 0.6
	//   neighbors of center: 0 + 0.1*(1.0 - 0) = 0.1
	//   all boundary cells: 0 (fixed, never updated)
	g := grid.New(3, 3)
	g.Set(1, 1, 1.0)
	result := ref.Solve(g, 1.0, 0.1, 1.0, 1)

	cases := []struct {
		r, c int
		want float64
	}{
		{1, 1, 0.6},
		{0, 1, 0.0}, // boundary, never updated
		{1, 0, 0.0}, // boundary, never updated
	}
	for _, tc := range cases {
		got := result.At(tc.r, tc.c)
		if math.Abs(got-tc.want) > 1e-14 {
			t.Errorf("(%d,%d): want %v, got %v", tc.r, tc.c, tc.want, got)
		}
	}
}

func TestRefBoundariesFixed(t *testing.T) {
	g := grid.New(4, 4)
	g.Set(2, 2, 1.0)
	result := ref.Solve(g, 0.1, 0.001, 0.25, 100)
	for c := 0; c < 4; c++ {
		if result.At(0, c) != 0 || result.At(3, c) != 0 {
			t.Errorf("boundary row modified at col %d", c)
		}
	}
	for r := 0; r < 4; r++ {
		if result.At(r, 0) != 0 || result.At(r, 3) != 0 {
			t.Errorf("boundary col modified at row %d", r)
		}
	}
}
