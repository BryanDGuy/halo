// internal/sim/sim.go
package sim

import (
	"context"
	"sync"

	"github.com/BryanDGuy/halo/internal/grid"
	"github.com/BryanDGuy/halo/internal/worker"
)

// Sim orchestrates a pw×pw grid of workers.
type Sim struct {
	workers    [][]*worker.Worker
	pw         int
	numWorkers int
	wg         sync.WaitGroup
}

// New creates a Sim from a pw×pw tile grid, wiring all inter-worker channels.
// Channels are buffered (cap 1) to prevent deadlock during halo exchange.
func New(tiles [][]*grid.Tile, alpha, dt, h float64) *Sim {
	pw := len(tiles)
	workers := make([][]*worker.Worker, pw)
	for i := range workers {
		workers[i] = make([]*worker.Worker, pw)
		for j := range workers[i] {
			workers[i][j] = worker.New(tiles[i][j], alpha, dt, h)
		}
	}

	// Wire vertical channels between row i (top) and row i+1 (bottom).
	for i := 0; i < pw-1; i++ {
		for j := 0; j < pw; j++ {
			southward := make(chan []float64, 1) // top sends south boundary downward
			northward := make(chan []float64, 1) // bottom sends north boundary upward
			workers[i][j].SetSendS(southward)
			workers[i][j].SetRecvS(northward)
			workers[i+1][j].SetRecvN(southward)
			workers[i+1][j].SetSendN(northward)
		}
	}

	// Wire horizontal channels between col j (left) and col j+1 (right).
	for i := 0; i < pw; i++ {
		for j := 0; j < pw-1; j++ {
			eastward := make(chan []float64, 1) // left sends east boundary rightward
			westward := make(chan []float64, 1) // right sends west boundary leftward
			workers[i][j].SetSendE(eastward)
			workers[i][j].SetRecvE(westward)
			workers[i][j+1].SetRecvW(eastward)
			workers[i][j+1].SetSendW(westward)
		}
	}

	return &Sim{workers: workers, pw: pw, numWorkers: pw * pw}
}

// Start launches all worker goroutines. Must be called before Step.
func (s *Sim) Start(ctx context.Context) {
	for _, row := range s.workers {
		for _, w := range row {
			go w.Run(ctx, &s.wg)
		}
	}
}

// Step runs one timestep. It returns false if ctx is already cancelled.
func (s *Sim) Step(ctx context.Context) bool {
	if ctx.Err() != nil {
		return false
	}
	s.wg.Add(s.numWorkers)
	for _, row := range s.workers {
		for _, w := range row {
			w.Start() <- struct{}{}
		}
	}
	s.wg.Wait()
	// Enforce Dirichlet BC: reset full-grid boundary cells that were updated
	// by StepInto back to 0, matching the ref solver which never updates them.
	s.fixBoundaries()
	return true
}

// fixBoundaries zeroes the full-grid boundary interior cells in edge tiles.
// Tile.StepInto updates all interior cells including those at the full-grid
// boundary; ref.Solve leaves those cells fixed at 0 (Dirichlet BC).
func (s *Sim) fixBoundaries() {
	for i, row := range s.workers {
		for j, w := range row {
			t := w.CurrentTile()
			if i == 0 {
				for c := 0; c < t.Cols; c++ {
					t.SetInteriorAt(0, c, 0)
				}
			}
			if i == s.pw-1 {
				for c := 0; c < t.Cols; c++ {
					t.SetInteriorAt(t.Rows-1, c, 0)
				}
			}
			if j == 0 {
				for r := 0; r < t.Rows; r++ {
					t.SetInteriorAt(r, 0, 0)
				}
			}
			if j == s.pw-1 {
				for r := 0; r < t.Rows; r++ {
					t.SetInteriorAt(r, t.Cols-1, 0)
				}
			}
		}
	}
}

// Collect copies the current tile state of every worker into g.
func (s *Sim) Collect(g *grid.Grid) {
	for _, row := range s.workers {
		for _, w := range row {
			grid.CollectTile(w.CurrentTile(), g)
		}
	}
}
