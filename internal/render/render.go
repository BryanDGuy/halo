package render

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BryanDGuy/halo/internal/grid"
)

var gradient = []rune{' ', '░', '▒', '▓', '█'}

func toChar(v, scale float64) rune {
	if scale <= 0 || v <= 0 {
		return gradient[0]
	}
	norm := v / scale
	if norm >= 1 {
		return gradient[len(gradient)-1]
	}
	idx := int(norm * float64(len(gradient)-1))
	return gradient[idx]
}

// FrameTo writes the heatmap to w, normalizing to the grid's current max value.
func FrameTo(w io.Writer, g *grid.Grid) {
	var peak float64
	for _, v := range g.Data {
		if v > peak {
			peak = v
		}
	}
	var sb strings.Builder
	for r := 0; r < g.Rows; r++ {
		for c := 0; c < g.Cols; c++ {
			sb.WriteRune(toChar(g.At(r, c), peak))
		}
		sb.WriteByte('\n')
	}
	fmt.Fprint(w, sb.String())
}

// Frame clears the terminal and renders g as a heatmap.
func Frame(g *grid.Grid) {
	fmt.Fprint(os.Stdout, "\033[H\033[2J") // ANSI: cursor home + clear screen
	FrameTo(os.Stdout, g)
}
