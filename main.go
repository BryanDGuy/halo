package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/BryanDGuy/halo/internal/grid"
	"github.com/BryanDGuy/halo/internal/render"
	"github.com/BryanDGuy/halo/internal/sim"
)

func main() {
	rows    := flag.Int("rows", 64, "grid height")
	cols    := flag.Int("cols", 64, "grid width")
	workers := flag.Int("workers", 2, "worker grid dimension (NxN)")
	alpha   := flag.Float64("alpha", 0.1, "thermal diffusivity")
	dt      := flag.Float64("dt", 0, "timestep (0 = auto)")
	steps   := flag.Int("steps", 0, "steps to run (0 = run until Ctrl+C)")
	flag.Parse()

	h := 1.0 / float64(max(*rows, *cols)) // built-in max, Go 1.21+
	if *dt == 0 {
		*dt = 0.24 * h * h / *alpha
	}
	if *alpha**dt/(h*h) > 0.25 {
		fmt.Fprintf(os.Stderr, "warning: unstable timestep (α·Δt/h²=%.4f > 0.25)\n",
			*alpha**dt/(h*h))
	}

	g := grid.New(*rows, *cols)
	g.Set(*rows/2, *cols/2, 1.0) // hotspot at center

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

	s.Start(ctx)

	display := grid.New(*rows, *cols)
	step := 0
	for {
		if *steps > 0 && step >= *steps {
			break
		}
		if !s.Step(ctx) {
			break
		}
		s.Collect(display)
		render.Frame(display)
		fmt.Printf("step %d  dt=%.6f  α=%.3f\n", step+1, *dt, *alpha)
		step++
	}
}
