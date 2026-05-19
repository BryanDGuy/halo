# Halo — Design Spec

**Date:** 2026-05-19  
**Module:** `github.com/BryanDGuy/halo`  
**Goal:** Working 2D heat equation simulation distributed across goroutines using the halo exchange pattern. Doubles as a concurrency learning exercise.

---

## Overview

The simulation discretizes the 2D heat equation onto a grid and evolves it forward in time using an explicit 5-point stencil. The grid is partitioned into 2D tiles, each owned by a dedicated goroutine (worker). Workers exchange boundary data (halos) with their neighbors via channels each timestep, then compute their tile's interior. A centralized orchestrator drives the step loop and renders an ASCII heatmap to the terminal.

Stability constraint (explicit scheme, 2D): `α·Δt/h² ≤ 0.25`, where `h = 1.0 / max(rows, cols)` (unit square domain).

---

## Package Structure

```
halo/
  main.go                  — flags, wires sim, handles SIGINT
  internal/
    grid/
      grid.go              — Grid type (NxM float64), 2D tile decomposition
      grid_test.go
    worker/
      worker.go            — Worker goroutine: tile ownership, halo exchange, stencil compute
      worker_test.go
    sim/
      sim.go               — Simulator: orchestrates workers, drives step loop
      sim_test.go
    render/
      render.go            — ASCII heatmap: float64 → block characters, terminal clear
    ref/
      ref.go               — Single-goroutine reference solver (test use only)
```

---

## Key Types

| Type | Package | Responsibility |
|------|---------|----------------|
| `Grid` | `grid` | NxM `float64` slice; `Decompose(rows, cols int) [][]Tile` |
| `Tile` | `grid` | Sub-grid with pre-allocated ghost border rows/cols for halos |
| `Worker` | `worker` | Owns a `Tile`; holds 4 buffered channel pairs (N/S/E/W send+recv) |
| `Sim` | `sim` | Holds `[][]Worker`, a tick channel, and a `sync.WaitGroup` |

---

## Concurrency Model

**Orchestrator-driven step loop.** The main goroutine (`Sim.Run`) drives each timestep:

```
// sim.Run() — main goroutine
for each step:
  1. close(tick)                 // broadcast: all workers unblock simultaneously
  2. tick = make(chan struct{})   // reset for next step
  3. wg.Wait()                   // block until all workers signal done
  4. render(grid)                // stitch tiles → ASCII heatmap

// worker.Run() — one goroutine per tile
for each step:
  1. <-tick                      // wait for orchestrator
  2. send boundaries → N/S/E/W  // non-blocking (buffered chan, cap 1)
  3. recv halos ← N/S/E/W       // blocks until all 4 neighbors have sent
  4. compute interior            // 5-point stencil
  5. wg.Done()                   // signal orchestrator
```

**Halo channels are buffered (cap 1).** This prevents the send/send deadlock that would occur if two adjacent workers on unbuffered channels both tried to send to each other before either was ready to receive.

**`close(tick)` broadcast.** Closing a channel unblocks all receivers simultaneously — cleaner than sending N individual signals.

**Rendering is race-free.** Workers' tiles are only read by the renderer after `wg.Wait()` confirms all workers have finished writing for the step.

---

## Grid Decomposition

The grid is split into a `P×Q` tile grid where `P = Q = -workers` (always square). Each tile is assigned a `(row, col)` index in the worker grid. Corner and edge tiles have fewer than 4 neighbors; channels for absent neighbors are nil and skipped. Ghost cells in the boundary direction (where no neighbor exists) are held fixed at 0.0, implementing Dirichlet boundary conditions on the outer edges of the full grid.

---

## Configuration (CLI Flags)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-rows` | int | 64 | Grid height in cells |
| `-cols` | int | 64 | Grid width in cells |
| `-workers` | int | 2 | Worker grid dimension (2 → 2×2 = 4 workers) |
| `-alpha` | float64 | 0.1 | Thermal diffusivity |
| `-dt` | float64 | auto | Timestep; auto-computed as `0.24 * h² / alpha` if omitted (just under stability limit) |
| `-steps` | int | 0 | Steps to run; 0 = run until Ctrl+C |

---

## Initial Conditions

A hot spot (value 1.0) in the center of the grid, everything else 0.0. Fixed boundary conditions (edges held at 0.0).

---

## ASCII Rendering

Values mapped to a gradient of Unicode block characters (`░▒▓█`) based on normalized temperature. Terminal is cleared each frame with an ANSI escape. Frame rate is not throttled — runs as fast as the simulation allows; users can add a `-delay` flag later if needed.

---

## Error Handling

- Invalid configuration (negative grid size, unstable dt): `log.Fatal` at startup.
- SIGINT (Ctrl+C): context cancellation propagated to `Sim.Run`; workers check `ctx.Done()` each step and exit cleanly.
- No recoverable errors in the simulation loop itself.

---

## Testing

### `grid/grid_test.go`
- `Decompose(4, 4)` on an 8×8 grid produces 16 tiles with correct bounds and no overlap.
- Ghost border rows/cols are allocated with the correct size.

### `worker/worker_test.go`
- Single step on a 2×2 worker grid: halo values received match what the neighbor sent.
- Stencil math on a known input produces the expected output.

### `sim/sim_test.go`
- Run N steps; verify the result matches the single-goroutine reference implementation (`internal/ref`) cell-for-cell.
- Run to steady state; verify max delta between steps falls below ε.

### `internal/ref/ref.go`
Single-goroutine solver implementing the same stencil. Used only in tests as the ground truth to diff against the concurrent implementation.
