package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BryanDGuy/halo/internal/grid"
	"github.com/BryanDGuy/halo/internal/render"
	"github.com/BryanDGuy/halo/internal/sim"
	"golang.org/x/term"
)

const framePeriod = time.Second / 30

func main() {
	termCols, termRows := termSize()

	// Terminal characters are ~2× taller than wide, so half as many rows
	// fills the same physical area, keeping the grid visually square.
	rows := min(termRows-1, termCols/2)
	cols := termCols

	workers := flag.Int("workers", 2, "worker grid dimension (NxN)")
	alpha := flag.Float64("alpha", 0.1, "thermal diffusivity")
	dt := flag.Float64("dt", 0, "timestep (0 = auto)")
	steps := flag.Int("steps", 0, "steps to run (0 = run until Ctrl+C)")
	flag.Parse()

	h := 1.0 / float64(max(rows, cols))
	if *dt == 0 {
		*dt = 0.24 * h * h / *alpha
	}
	if *alpha**dt/(h*h) > 0.25 {
		fmt.Fprintln(os.Stderr, "warning: unstable timestep (α·Δt/h² > 0.25) — reduce -dt or -alpha")
	}

	g := grid.New(rows, cols)
	g.Set(rows/2, cols/2, 1.0)

	tiles := grid.Decompose(g, *workers)
	s := sim.New(tiles, *alpha, *dt, h)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	render.Init()
	defer render.Cleanup()

	s.Start(ctx)

	display := grid.New(rows, cols)
	step := 0
	for *steps == 0 || step < *steps {
		start := time.Now()
		if !s.Step(ctx) {
			break
		}
		s.Collect(display)
		render.Frame(display, fmt.Sprintf("step %-6d  dt=%.6f  α=%.3f", step+1, *dt, *alpha))
		step++
		if elapsed := time.Since(start); elapsed < framePeriod {
			time.Sleep(framePeriod - elapsed)
		}
	}
}

// termSize returns the terminal dimensions, falling back to 80×24 if unavailable.
func termSize() (cols, rows int) {
	fd := os.Stdout.Fd()
	// File descriptors are small non-negative integers on all supported platforms;
	// the bounds check satisfies static analysis of the uintptr→int conversion.
	if fd > uintptr(math.MaxInt) {
		return 80, 24
	}
	cols, rows, err := term.GetSize(int(fd))
	if err != nil || cols <= 0 || rows <= 0 {
		return 80, 24
	}
	return cols, rows
}
