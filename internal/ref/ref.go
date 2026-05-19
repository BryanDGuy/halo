// internal/ref/ref.go
package ref

import "github.com/BryanDGuy/halo/internal/grid"

// Solve runs n steps of the explicit 5-point stencil on g and returns the result.
// Boundary cells (row 0, row Rows-1, col 0, col Cols-1) are held fixed at their
// initial values (Dirichlet BC). g is not modified.
func Solve(g *grid.Grid, alpha, dt, h float64, n int) *grid.Grid {
	rows, cols := g.Rows, g.Cols
	cur := grid.New(rows, cols)
	copy(cur.Data, g.Data)
	nxt := grid.New(rows, cols)
	r := alpha * dt / (h * h)
	for step := 0; step < n; step++ {
		for i := 1; i < rows-1; i++ {
			for j := 1; j < cols-1; j++ {
				c := cur.At(i, j)
				nxt.Set(i, j, c+r*(cur.At(i-1, j)+cur.At(i+1, j)+cur.At(i, j-1)+cur.At(i, j+1)-4*c))
			}
		}
		cur, nxt = nxt, cur
		// Clear interior of nxt for next iteration (boundaries stay 0).
		for i := 1; i < rows-1; i++ {
			for j := 1; j < cols-1; j++ {
				nxt.Set(i, j, 0)
			}
		}
	}
	return cur
}
