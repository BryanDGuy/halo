# halo

A distributed stencil solver in Go. It splits a grid across workers, each of which
computes the interior of its tile and exchanges boundary cells — the *halo* — with
its neighbors every timestep. The reference simulation is the 2D heat equation,
rendered live in the terminal as an ASCII heatmap.

## The idea

The 2D heat equation says a point's temperature changes based only on its immediate
neighbors:

```
∂u/∂t = α (∂²u/∂x² + ∂²u/∂y²)
```

Discretized onto a grid, each cell's next value is a nudge toward the average of its
four neighbors (the 5-point stencil):

```
u_new[i][j] = u[i][j] + (α·Δt/h²) · ( u[i+1][j] + u[i-1][j]
                                     + u[i][j+1] + u[i][j-1]
                                     − 4·u[i][j] )
```

Because the update is local, the grid can be tiled across workers. The only data a
worker needs from outside its tile is the row/column of cells just past its edges —
the halo. Each step is: exchange halos → compute interior → barrier → repeat.

Stability constraint (explicit scheme, 2D): `α·Δt/h² ≤ 1/4`.

## Running

```
make build
./bin/halo
```

Press `Ctrl+C` to exit.

**Options:**

| Flag | Default | Description |
|------|---------|-------------|
| `-rows` | 64 | Grid height |
| `-cols` | 64 | Grid width |
| `-workers` | 2 | Worker grid dimension (NxN) |
| `-alpha` | 0.1 | Thermal diffusivity |
| `-dt` | auto | Timestep (0 = auto-compute stable value) |
| `-steps` | 0 | Steps to run (0 = run until Ctrl+C) |

Example — larger grid with a 4×4 worker pool:

```
./bin/halo -rows 80 -cols 80 -workers 4
```

## Development

```
make test        # run tests with race detector
make lint        # golangci-lint
make fmt         # gofumpt
make vet         # go vet
make tidy        # go mod tidy
make modernize   # check for modernizable patterns
make modernize-fix  # apply modernize fixes
make update-deps # bump all dependencies
```

## Architecture

```
main.go          orchestrates flags, signal handling, render loop
internal/
  grid/          Grid and Tile types; Decompose, CollectTile, StepInto
  worker/        per-tile goroutine; halo exchange via buffered channels
  sim/           Sim — wires workers, drives step loop, enforces Dirichlet BC
  render/        ASCII heatmap renderer; in-place terminal animation
  ref/           single-goroutine reference solver (used in tests)
```

Each worker goroutine owns a tile and communicates with up to four neighbors through
buffered channels. The `Sim` orchestrator sends a start signal to every worker each
step, then waits on a `WaitGroup`. After all workers finish, `Sim` enforces Dirichlet
boundary conditions (outer edges fixed at 0) before the next step.
