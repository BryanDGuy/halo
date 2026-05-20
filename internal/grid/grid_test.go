package grid_test

import (
	"testing"

	"github.com/BryanDGuy/halo/internal/grid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecompose2x2(t *testing.T) {
	g := grid.New(8, 8)
	tiles := grid.Decompose(g, 2)
	require.Len(t, tiles, 2)
	require.Len(t, tiles[0], 2)
	for i := range tiles {
		for j, tile := range tiles[i] {
			assert.Equal(t, 4, tile.Rows, "tiles[%d][%d].Rows", i, j)
			assert.Equal(t, 4, tile.Cols, "tiles[%d][%d].Cols", i, j)
		}
	}
	assert.Equal(t, 0, tiles[0][0].RowStart)
	assert.Equal(t, 0, tiles[0][0].ColStart)
	assert.Equal(t, 4, tiles[1][1].RowStart)
	assert.Equal(t, 4, tiles[1][1].ColStart)
}

func TestDecomposeNonDivisible(t *testing.T) {
	// splitEvenly(9,2) = [5,4]
	g := grid.New(9, 9)
	tiles := grid.Decompose(g, 2)
	assert.Equal(t, 5, tiles[0][0].Rows)
	assert.Equal(t, 4, tiles[1][0].Rows)
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
	assert.Equal(t, g.Data, g2.Data)
}

func TestNorthSouthBoundaries(t *testing.T) {
	g := grid.New(4, 4)
	for r := range 4 {
		for c := range 4 {
			g.Set(r, c, float64(r*4+c))
		}
	}
	tiles := grid.Decompose(g, 2)
	t00 := tiles[0][0] // interior rows 0-1, cols 0-1 of full grid

	// NorthBoundary = top interior row = grid row 0 = [0, 1]
	assert.Equal(t, []float64{0, 1}, t00.NorthBoundary())
	// SouthBoundary = bottom interior row = grid row 1 = [4, 5]
	assert.Equal(t, []float64{4, 5}, t00.SouthBoundary())
}

func TestEastWestBoundaries(t *testing.T) {
	g := grid.New(4, 4)
	for r := range 4 {
		for c := range 4 {
			g.Set(r, c, float64(r*4+c))
		}
	}
	tiles := grid.Decompose(g, 2)
	t00 := tiles[0][0] // interior rows 0-1, cols 0-1 of full grid

	// WestBoundary = leftmost interior col = grid col 0 = [0, 4]
	assert.Equal(t, []float64{0, 4}, t00.WestBoundary())
	// EastBoundary = rightmost interior col = grid col 1 = [1, 5]
	assert.Equal(t, []float64{1, 5}, t00.EastBoundary())
}

func TestStepIntoUsesNorthGhost(t *testing.T) {
	// 1x1 interior tile; inject north ghost and verify stencil uses it.
	// alpha=1, dt=0.1, h=1 → r=0.1
	// center=0, north ghost=10, all others=0
	// new center = 0 + 0.1*(10+0+0+0 - 4*0) = 1.0
	g := grid.New(1, 1)
	tiles := grid.Decompose(g, 1)
	tile := tiles[0][0]
	tile.SetNorthGhost([]float64{10.0})

	nxt := grid.NewTile(0, 0, 1, 1)
	tile.StepInto(nxt, 1.0, 0.1, 1.0)

	assert.InDelta(t, 1.0, nxt.InteriorAt(0, 0), 1e-14)
}

func TestStepIntoUsesSouthGhost(t *testing.T) {
	// center=0, south ghost=10, all others=0 → new center = 1.0
	g := grid.New(1, 1)
	tiles := grid.Decompose(g, 1)
	tile := tiles[0][0]
	tile.SetSouthGhost([]float64{10.0})

	nxt := grid.NewTile(0, 0, 1, 1)
	tile.StepInto(nxt, 1.0, 0.1, 1.0)

	assert.InDelta(t, 1.0, nxt.InteriorAt(0, 0), 1e-14)
}

func TestStepIntoUsesWestGhost(t *testing.T) {
	// center=0, west ghost=10, all others=0 → new center = 1.0
	g := grid.New(1, 1)
	tiles := grid.Decompose(g, 1)
	tile := tiles[0][0]
	tile.SetWestGhost([]float64{10.0})

	nxt := grid.NewTile(0, 0, 1, 1)
	tile.StepInto(nxt, 1.0, 0.1, 1.0)

	assert.InDelta(t, 1.0, nxt.InteriorAt(0, 0), 1e-14)
}

func TestStepIntoUsesEastGhost(t *testing.T) {
	// center=0, east ghost=10, all others=0 → new center = 1.0
	g := grid.New(1, 1)
	tiles := grid.Decompose(g, 1)
	tile := tiles[0][0]
	tile.SetEastGhost([]float64{10.0})

	nxt := grid.NewTile(0, 0, 1, 1)
	tile.StepInto(nxt, 1.0, 0.1, 1.0)

	assert.InDelta(t, 1.0, nxt.InteriorAt(0, 0), 1e-14)
}
