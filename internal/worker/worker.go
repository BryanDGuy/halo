// internal/worker/worker.go
package worker

import (
	"context"
	"sync"

	"github.com/BryanDGuy/halo/internal/grid"
)

// Worker owns a tile and communicates with up to 4 neighbors via buffered channels.
// Channels are nil for missing neighbors (edge/corner tiles); nil directions are
// skipped during exchange, leaving ghost cells at their initial value of 0 (Dirichlet BC).
type Worker struct {
	cur, nxt     *grid.Tile
	alpha, dt, h float64
	startCh      chan struct{}
	sendN, sendS chan<- []float64
	sendE, sendW chan<- []float64
	recvN, recvS <-chan []float64
	recvE, recvW <-chan []float64
}

func New(tile *grid.Tile, alpha, dt, h float64) *Worker {
	return &Worker{
		cur:     tile,
		nxt:     grid.NewTile(tile.RowStart, tile.ColStart, tile.Rows, tile.Cols),
		alpha:   alpha,
		dt:      dt,
		h:       h,
		startCh: make(chan struct{}),
	}
}

// Start returns the channel the Sim uses to trigger each step.
func (w *Worker) Start() chan<- struct{} { return w.startCh }

// CurrentTile returns the tile holding the latest computed state.
func (w *Worker) CurrentTile() *grid.Tile { return w.cur }

// SetSendN / SetRecvN etc. are called by Sim during wiring.
func (w *Worker) SetSendN(ch chan<- []float64) { w.sendN = ch }
func (w *Worker) SetSendS(ch chan<- []float64) { w.sendS = ch }
func (w *Worker) SetSendE(ch chan<- []float64) { w.sendE = ch }
func (w *Worker) SetSendW(ch chan<- []float64) { w.sendW = ch }
func (w *Worker) SetRecvN(ch <-chan []float64) { w.recvN = ch }
func (w *Worker) SetRecvS(ch <-chan []float64) { w.recvS = ch }
func (w *Worker) SetRecvE(ch <-chan []float64) { w.recvE = ch }
func (w *Worker) SetRecvW(ch <-chan []float64) { w.recvW = ch }

// Run is the goroutine entry point. It waits for a start signal each step,
// exchanges halos, runs the stencil, then calls wg.Done().
func (w *Worker) Run(ctx context.Context, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.startCh:
		}
		w.exchangeHalos()
		w.cur.StepInto(w.nxt, w.alpha, w.dt, w.h)
		w.cur, w.nxt = w.nxt, w.cur
		wg.Done()
	}
}

// exchangeHalos sends boundary rows/cols to neighbors and receives their halos.
// Sends happen first into buffered channels (non-blocking), then receives block
// until all neighbors have sent — this prevents deadlock without ordering constraints.
func (w *Worker) exchangeHalos() {
	if w.sendN != nil {
		w.sendN <- w.cur.NorthBoundary()
	}
	if w.sendS != nil {
		w.sendS <- w.cur.SouthBoundary()
	}
	if w.sendE != nil {
		w.sendE <- w.cur.EastBoundary()
	}
	if w.sendW != nil {
		w.sendW <- w.cur.WestBoundary()
	}

	if w.recvN != nil {
		w.cur.SetNorthGhost(<-w.recvN)
	}
	if w.recvS != nil {
		w.cur.SetSouthGhost(<-w.recvS)
	}
	if w.recvE != nil {
		w.cur.SetEastGhost(<-w.recvE)
	}
	if w.recvW != nil {
		w.cur.SetWestGhost(<-w.recvW)
	}
}
