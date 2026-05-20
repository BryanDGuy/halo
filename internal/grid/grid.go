// internal/grid/grid.go
package grid

// Grid is an N×M grid of float64 values.
type Grid struct {
	Rows, Cols int
	Data       []float64
}

func New(rows, cols int) *Grid {
	return &Grid{Rows: rows, Cols: cols, Data: make([]float64, rows*cols)}
}

func (g *Grid) At(r, c int) float64     { return g.Data[r*g.Cols+c] }
func (g *Grid) Set(r, c int, v float64) { g.Data[r*g.Cols+c] = v }

// Tile is a sub-region of the grid owned by a worker.
// Data is (Rows+2)×(Cols+2): row 0 and row Rows+1 are north/south ghost rows;
// col 0 and col Cols+1 are west/east ghost columns. Boundary ghosts default to 0
// (Dirichlet BC).
type Tile struct {
	RowStart, ColStart int
	Rows, Cols         int
	Data               []float64
	tc                 int // total cols = Cols+2
}

func NewTile(rowStart, colStart, rows, cols int) *Tile {
	tc := cols + 2
	return &Tile{
		RowStart: rowStart, ColStart: colStart,
		Rows: rows, Cols: cols,
		Data: make([]float64, (rows+2)*tc),
		tc:   tc,
	}
}

func (t *Tile) at(r, c int) float64     { return t.Data[r*t.tc+c] }
func (t *Tile) set(r, c int, v float64) { t.Data[r*t.tc+c] = v }

// InteriorAt returns the value at interior position (r, c), 0-indexed.
func (t *Tile) InteriorAt(r, c int) float64 { return t.at(r+1, c+1) }

// SetInteriorAt sets the value at interior position (r, c), 0-indexed.
func (t *Tile) SetInteriorAt(r, c int, v float64) { t.set(r+1, c+1, v) }

// NorthBoundary returns a copy of the topmost interior row (sent to north neighbor).
func (t *Tile) NorthBoundary() []float64 {
	row := make([]float64, t.Cols)
	for c := 0; c < t.Cols; c++ {
		row[c] = t.at(1, c+1)
	}
	return row
}

// SouthBoundary returns a copy of the bottommost interior row (sent to south neighbor).
func (t *Tile) SouthBoundary() []float64 {
	row := make([]float64, t.Cols)
	for c := 0; c < t.Cols; c++ {
		row[c] = t.at(t.Rows, c+1)
	}
	return row
}

// WestBoundary returns a copy of the leftmost interior column (sent to west neighbor).
func (t *Tile) WestBoundary() []float64 {
	col := make([]float64, t.Rows)
	for r := 0; r < t.Rows; r++ {
		col[r] = t.at(r+1, 1)
	}
	return col
}

// EastBoundary returns a copy of the rightmost interior column (sent to east neighbor).
func (t *Tile) EastBoundary() []float64 {
	col := make([]float64, t.Rows)
	for r := 0; r < t.Rows; r++ {
		col[r] = t.at(r+1, t.Cols)
	}
	return col
}

// SetNorthGhost writes vals into the north ghost row (row 0).
func (t *Tile) SetNorthGhost(vals []float64) {
	for c, v := range vals {
		t.set(0, c+1, v)
	}
}

// SetSouthGhost writes vals into the south ghost row (row Rows+1).
func (t *Tile) SetSouthGhost(vals []float64) {
	for c, v := range vals {
		t.set(t.Rows+1, c+1, v)
	}
}

// SetWestGhost writes vals into the west ghost column (col 0).
func (t *Tile) SetWestGhost(vals []float64) {
	for r, v := range vals {
		t.set(r+1, 0, v)
	}
}

// SetEastGhost writes vals into the east ghost column (col Cols+1).
func (t *Tile) SetEastGhost(vals []float64) {
	for r, v := range vals {
		t.set(r+1, t.Cols+1, v)
	}
}

// StepInto computes one explicit timestep of the 5-point stencil from t into dst.
// dst must have the same dimensions as t.
func (t *Tile) StepInto(dst *Tile, alpha, dt, h float64) {
	r := alpha * dt / (h * h)
	for i := 1; i <= t.Rows; i++ {
		for j := 1; j <= t.Cols; j++ {
			center := t.at(i, j)
			dst.set(i, j, center+r*(t.at(i-1, j)+t.at(i+1, j)+t.at(i, j-1)+t.at(i, j+1)-4*center))
		}
	}
}

// Decompose splits g into a pw×pw tile grid. Tile sizes are as even as possible;
// if rows/cols are not divisible by pw, earlier tiles get the extra cells.
func Decompose(g *Grid, pw int) [][]*Tile {
	rowH := splitEvenly(g.Rows, pw)
	colW := splitEvenly(g.Cols, pw)
	tiles := make([][]*Tile, pw)
	rowStart := 0
	for i := range tiles {
		tiles[i] = make([]*Tile, pw)
		colStart := 0
		for j := range tiles[i] {
			t := NewTile(rowStart, colStart, rowH[i], colW[j])
			for r := 0; r < rowH[i]; r++ {
				for c := 0; c < colW[j]; c++ {
					t.set(r+1, c+1, g.At(rowStart+r, colStart+c))
				}
			}
			tiles[i][j] = t
			colStart += colW[j]
		}
		rowStart += rowH[i]
	}
	return tiles
}

// CollectTile copies the interior of t back into g.
func CollectTile(t *Tile, g *Grid) {
	for r := 0; r < t.Rows; r++ {
		for c := 0; c < t.Cols; c++ {
			g.Set(t.RowStart+r, t.ColStart+c, t.at(r+1, c+1))
		}
	}
}

func splitEvenly(n, p int) []int {
	parts := make([]int, p)
	base, rem := n/p, n%p
	for i := range parts {
		parts[i] = base
		if i < rem {
			parts[i]++
		}
	}
	return parts
}
