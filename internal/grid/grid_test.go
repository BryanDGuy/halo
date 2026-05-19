// internal/grid/grid_test.go
package grid_test

import (
	"math"
	"testing"

	"github.com/BryanDGuy/halo/internal/grid"
)

func TestDecompose2x2(t *testing.T) {
	g := grid.New(8, 8)
	tiles := grid.Decompose(g, 2)
	if len(tiles) != 2 || len(tiles[0]) != 2 {
		t.Fatalf("want 2x2, got %dx%d", len(tiles), len(tiles[0]))
	}
	for i := range tiles {
		for j, tile := range tiles[i] {
			if tile.Rows != 4 || tile.Cols != 4 {
				t.Errorf("tiles[%d][%d]: want 4x4, got %dx%d", i, j, tile.Rows, tile.Cols)
			}
		}
	}
	if tiles[0][0].RowStart != 0 || tiles[0][0].ColStart != 0 {
		t.Errorf("tiles[0][0]: want start (0,0), got (%d,%d)", tiles[0][0].RowStart, tiles[0][0].ColStart)
	}
	if tiles[1][1].RowStart != 4 || tiles[1][1].ColStart != 4 {
		t.Errorf("tiles[1][1]: want start (4,4), got (%d,%d)", tiles[1][1].RowStart, tiles[1][1].ColStart)
	}
}

func TestDecomposeNonDivisible(t *testing.T) {
	g := grid.New(9, 9)
	tiles := grid.Decompose(g, 2)
	// splitEvenly(9,2) = [5,4]
	if tiles[0][0].Rows != 5 {
		t.Errorf("tiles[0][0].Rows: want 5, got %d", tiles[0][0].Rows)
	}
	if tiles[1][0].Rows != 4 {
		t.Errorf("tiles[1][0].Rows: want 4, got %d", tiles[1][0].Rows)
	}
}

func TestCollectTileRoundTrip(t *testing.T) {
	g := grid.New(4, 4)
	for i := range g.Data {
		g.Data[i] = float64(i + 1)
	}
	tiles := grid.Decompose(g, 2)
	g2 := grid.New(4, 4)
	for _, row := range tiles {
		for _, tile := range row {
			grid.CollectTile(tile, g2)
		}
	}
	for i := range g.Data {
		if g.Data[i] != g2.Data[i] {
			t.Errorf("cell %d: want %v, got %v", i, g.Data[i], g2.Data[i])
		}
	}
}

func TestNorthSouthBoundaries(t *testing.T) {
	g := grid.New(4, 4)
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			g.Set(r, c, float64(r*4+c))
		}
	}
	tiles := grid.Decompose(g, 2)
	t00 := tiles[0][0] // interior rows 0-1, cols 0-1 of full grid

	// NorthBoundary = top interior row = grid row 0 = [0, 1]
	nb := t00.NorthBoundary()
	if nb[0] != 0 || nb[1] != 1 {
		t.Errorf("NorthBoundary: want [0 1], got %v", nb)
	}
	// SouthBoundary = bottom interior row = grid row 1 = [4, 5]
	sb := t00.SouthBoundary()
	if sb[0] != 4 || sb[1] != 5 {
		t.Errorf("SouthBoundary: want [4 5], got %v", sb)
	}
}

func TestStepIntoUsesGhost(t *testing.T) {
	// 1x1 interior tile; inject ghost values and verify stencil uses them.
	// alpha=1, dt=0.1, h=1 → r=0.1
	// center=0, north ghost=10, south ghost=0, east ghost=0, west ghost=0
	// new center = 0 + 0.1*(10+0+0+0 - 4*0) = 1.0
	g := grid.New(1, 1) // single interior cell, value 0
	tiles := grid.Decompose(g, 1)
	tile := tiles[0][0]
	tile.SetNorthGhost([]float64{10.0})

	nxt := grid.NewTile(0, 0, 1, 1)
	tile.StepInto(nxt, 1.0, 0.1, 1.0)

	got := nxt.InteriorAt(0, 0)
	if math.Abs(got-1.0) > 1e-14 {
		t.Errorf("want 1.0, got %v", got)
	}
}

func TestEastWestBoundaries(t *testing.T) {
	g := grid.New(4, 4)
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			g.Set(r, c, float64(r*4+c))
		}
	}
	tiles := grid.Decompose(g, 2)
	t00 := tiles[0][0] // interior rows 0-1, cols 0-1 of full grid

	// WestBoundary = leftmost interior col = grid col 0 = [0, 4]
	wb := t00.WestBoundary()
	if wb[0] != 0 || wb[1] != 4 {
		t.Errorf("WestBoundary: want [0 4], got %v", wb)
	}
	// EastBoundary = rightmost interior col = grid col 1 = [1, 5]
	eb := t00.EastBoundary()
	if eb[0] != 1 || eb[1] != 5 {
		t.Errorf("EastBoundary: want [1 5], got %v", eb)
	}
}

func TestStepIntoUsesWestGhost(t *testing.T) {
	// 1x1 interior tile; inject west ghost and verify stencil uses it.
	// alpha=1, dt=0.1, h=1 → r=0.1
	// center=0, west ghost=10, all others=0
	// new center = 0 + 0.1*(0+0+10+0 - 4*0) = 1.0
	g := grid.New(1, 1)
	tiles := grid.Decompose(g, 1)
	tile := tiles[0][0]
	tile.SetWestGhost([]float64{10.0})

	nxt := grid.NewTile(0, 0, 1, 1)
	tile.StepInto(nxt, 1.0, 0.1, 1.0)

	got := nxt.InteriorAt(0, 0)
	if math.Abs(got-1.0) > 1e-14 {
		t.Errorf("want 1.0, got %v", got)
	}
}

func TestStepIntoUsesSouthGhost(t *testing.T) {
	// alpha=1, dt=0.1, h=1 → r=0.1
	// center=0, south ghost=10, all others=0
	// new center = 0 + 0.1*(0+10+0+0 - 4*0) = 1.0
	g := grid.New(1, 1)
	tiles := grid.Decompose(g, 1)
	tile := tiles[0][0]
	tile.SetSouthGhost([]float64{10.0})

	nxt := grid.NewTile(0, 0, 1, 1)
	tile.StepInto(nxt, 1.0, 0.1, 1.0)

	got := nxt.InteriorAt(0, 0)
	if math.Abs(got-1.0) > 1e-14 {
		t.Errorf("want 1.0, got %v", got)
	}
}

func TestStepIntoUsesEastGhost(t *testing.T) {
	// alpha=1, dt=0.1, h=1 → r=0.1
	// center=0, east ghost=10, all others=0
	// new center = 0 + 0.1*(0+0+0+10 - 4*0) = 1.0
	g := grid.New(1, 1)
	tiles := grid.Decompose(g, 1)
	tile := tiles[0][0]
	tile.SetEastGhost([]float64{10.0})

	nxt := grid.NewTile(0, 0, 1, 1)
	tile.StepInto(nxt, 1.0, 0.1, 1.0)

	got := nxt.InteriorAt(0, 0)
	if math.Abs(got-1.0) > 1e-14 {
		t.Errorf("want 1.0, got %v", got)
	}
}
