// internal/worker/worker_test.go
package worker_test

import (
	"context"
	"sync"
	"testing"

	"github.com/BryanDGuy/halo/internal/grid"
	"github.com/BryanDGuy/halo/internal/worker"
)

func TestWorkerRunOneStep(t *testing.T) {
	// Single isolated worker (no neighbors), center cell = 1.0.
	// alpha=1, dt=0.1, h=1 → r=0.1
	// After one step the center of a 3x3 tile should be 0.6.
	g := grid.New(3, 3)
	g.Set(1, 1, 1.0)
	tiles := grid.Decompose(g, 1)

	w := worker.New(tiles[0][0], 1.0, 0.1, 1.0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go w.Run(ctx, &wg)

	w.Start() <- struct{}{} // fire one step
	wg.Wait()

	got := w.CurrentTile().InteriorAt(1, 1) // center interior cell (0-indexed)
	want := 0.6
	if got < want-1e-12 || got > want+1e-12 {
		t.Errorf("center after one step: want %v, got %v", want, got)
	}
}
